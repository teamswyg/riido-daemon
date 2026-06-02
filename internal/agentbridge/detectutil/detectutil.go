// Package detectutil owns the small C4 helper surface concrete provider
// adapters use to implement Detect: PATH lookup with env overrides and a
// short-running version probe.
//
// It does not own provider capability classification, provider-specific
// parsing, process spawning for runs, or scheduling decisions. We isolate
// it here so the agentbridge port doesn't grow an `os/exec` dependency
// and each provider package stays focused on its own command-line shape.
package detectutil

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const versionProbeTimeout = 10 * time.Second

// ResolveExecutable returns the absolute path to the binary.
//
// The override behaves as a PIN, not a hint:
//
//  1. If envOverride is non-empty and points to a regular file →
//     use it.
//  2. If envOverride is non-empty but the path does NOT exist or is a
//     directory → ("", false). We do NOT silently fall back to PATH;
//     doing so would let a misconfigured RIIDO_*_PATH appear to "work"
//     against a different binary than the operator intended.
//  3. If envOverride is empty → exec.LookPath(name).
//
// Returns (path, true) on success, ("", false) if not found.
func ResolveExecutable(name, envOverride string) (string, bool) {
	candidates := ResolveExecutableCandidates(name, envOverride)
	if len(candidates) == 0 {
		return "", false
	}
	return candidates[0], true
}

// ResolveExecutableCandidates returns executable candidates in PATH order.
//
// It preserves the same env override pin semantics as ResolveExecutable:
// an override returns at most that one file, and an invalid override
// returns no candidates. With no override, the first element matches the
// normal exec.LookPath result, followed by any later same-name PATH hits.
func ResolveExecutableCandidates(name, envOverride string) []string {
	override := strings.TrimSpace(envOverride)
	if override != "" {
		if !isRegularFile(override) {
			return nil
		}
		return []string{override}
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	var candidates []string
	seen := map[string]struct{}{}
	add := func(p string) {
		if p == "" {
			return
		}
		key := filepath.Clean(p)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		candidates = append(candidates, p)
	}

	p, err := exec.LookPath(name)
	if err == nil {
		add(p)
	}

	if strings.ContainsAny(name, `/\`) {
		return candidates
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		for _, candidateName := range executableNames(name) {
			path := filepath.Join(dir, candidateName)
			if isExecutableFile(path) {
				add(path)
			}
		}
	}

	return candidates
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode().IsRegular()
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || !info.Mode().IsRegular() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

func executableNames(name string) []string {
	if runtime.GOOS != "windows" || filepath.Ext(name) != "" {
		return []string{name}
	}

	exts := filepath.SplitList(os.Getenv("PATHEXT"))
	if len(exts) == 0 {
		exts = []string{".COM", ".EXE", ".BAT", ".CMD"}
	}
	out := make([]string, 0, len(exts))
	for _, ext := range exts {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			continue
		}
		out = append(out, name+ext)
	}
	if len(out) == 0 {
		return []string{name}
	}
	return out
}

// EnvOverride reads an env var by key, defaulting to os.Getenv.
// Provided for test injection — tests pass their own getter.
type EnvOverride func(key string) string

// OSEnv is the production env getter.
func OSEnv(key string) string { return os.Getenv(key) }

// VersionProbe runs `<exe> <args...>` with a short timeout and returns
// the trimmed combined stdout+stderr. Used by adapters to read
// `--version`. Errors (timeout, non-zero exit, missing binary) collapse
// to ("", false) so the caller decides whether to mark the runtime
// unavailable or just unknown.
//
// VersionProbe is lenient: a non-zero exit with any output is reported
// as `(output, true)`. Most provider `--version` flags exit 0 and emit
// a clean line, so this works fine for Claude / Codex / Cursor.
// Callers that MUST distinguish "successful version output" from
// "command failed with stderr" (e.g. OpenClaw, whose `--version` may
// emit a multi-line Node-dependency error when Node is too old) should
// use VersionProbeStrict instead — see audit M-8.
func VersionProbe(ctx context.Context, exe string, args ...string) (string, bool) {
	if exe == "" {
		return "", false
	}
	probeCtx, cancel := context.WithTimeout(ctx, versionProbeTimeout)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, exe, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		if errors.Is(probeCtx.Err(), context.DeadlineExceeded) {
			return "", false
		}
		// Some CLIs print version then exit non-zero on --version; we
		// still return what we captured if there's any output.
		if buf.Len() == 0 {
			return "", false
		}
	}
	return strings.TrimSpace(buf.String()), true
}

// ProbeResult is the output of VersionProbeStrict. It separates
// "command ran to completion" (OK=true) from "command was successful"
// (ExitCode == 0). Callers can refuse to extract a version from output
// when the command itself failed — see audit M-8 / openclaw.Detect.
type ProbeResult struct {
	// Output is the trimmed combined stdout+stderr.
	Output string
	// ExitCode is the process exit code. Meaningful only when OK is true.
	ExitCode int
	// OK reports whether the command actually ran. False when the
	// binary is missing, the context deadline was exceeded, or the
	// process was killed by a signal we can't classify as a normal
	// exit code.
	OK bool
}

// VersionProbeStrict runs the same probe as VersionProbe but exposes
// the exit code so callers can fail closed when the CLI reports an
// error through a non-zero exit.
func VersionProbeStrict(ctx context.Context, exe string, args ...string) ProbeResult {
	if exe == "" {
		return ProbeResult{OK: false}
	}
	probeCtx, cancel := context.WithTimeout(ctx, versionProbeTimeout)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, exe, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if errors.Is(probeCtx.Err(), context.DeadlineExceeded) {
		return ProbeResult{OK: false}
	}
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			// Failed to start, or killed by a signal we can't classify.
			return ProbeResult{OK: false}
		}
	}
	return ProbeResult{
		Output:   strings.TrimSpace(buf.String()),
		ExitCode: exitCode,
		OK:       true,
	}
}
