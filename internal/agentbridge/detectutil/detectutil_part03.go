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
)

// nodeVersionManagerBins resolves the bin dirs of node version managers, where
// npm-global CLIs (claude, openclaw) live under the active node version.
func nodeVersionManagerBins(home string) []string {
	globs := []string{
		filepath.Join(home, ".nvm", "versions", "node", "*", "bin"),
		filepath.Join(home, ".fnm", "node-versions", "*", "installation", "bin"),
		filepath.Join(home, "Library", "Application Support", "fnm", "node-versions", "*", "installation", "bin"),
		filepath.Join(home, ".local", "share", "fnm", "node-versions", "*", "installation", "bin"),
		filepath.Join(home, ".asdf", "installs", "nodejs", "*", "bin"),
	}
	var out []string
	for _, g := range globs {
		matches, err := filepath.Glob(g)
		if err != nil {
			continue
		}
		out = append(out, matches...)
	}
	return out
}

func windowsWellKnownDirs() []string {
	var dirs []string
	if appData := strings.TrimSpace(os.Getenv("APPDATA")); appData != "" {
		dirs = append(dirs, filepath.Join(appData, "npm"))
	}
	home, err := userHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return dirs
	}
	dirs = append(
		dirs,
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, ".cursor", "bin"),
		filepath.Join(home, ".claude", "bin"),
	)
	if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
		dirs = append(dirs, filepath.Join(localAppData, "Programs"))
	}
	return dirs
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
