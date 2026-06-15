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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const versionProbeTimeout = 10 * time.Second

// loginShellPATHTimeout bounds the one-shot login-shell PATH probe so a slow
// or misbehaving shell profile can never hang Detect.
const loginShellPATHTimeout = 3 * time.Second

// Markers wrap $PATH in the login-shell probe so noise printed by profile
// scripts (banners, version managers) is excluded from the captured value.
const (
	loginPATHMarkerStart = "__RIIDO_PATH_START__"
	loginPATHMarkerEnd   = "__RIIDO_PATH_END__"
)

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

	for _, dir := range augmentedSearchDirs() {
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

// augmentedSearchDirs is the seam tests override; production resolution uses
// productionSearchDirs.
var augmentedSearchDirs = productionSearchDirs

// productionSearchDirs returns the ordered, de-duplicated directories to scan
// for an executable.
//
// A daemon launched by the Riido Desktop app (or launchd/systemd) inherits a
// minimal PATH — on macOS launchd typically only /usr/bin:/bin:/usr/sbin:/sbin
// — which omits the Homebrew and per-user directories where provider CLIs
// (claude, codex, cursor-agent, openclaw) are actually installed. Resolving
// from the process PATH alone therefore reports installed providers as
// missing. We append the user's login-shell PATH (so resolution matches what
// the operator sees in a terminal) and a set of well-known install locations.
//
// Process PATH is listed first so an operator's explicit PATH still wins.
func productionSearchDirs() []string {
	var out []string
	seen := map[string]struct{}{}
	add := func(dir string) {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			return
		}
		key := filepath.Clean(dir)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, dir)
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		add(dir)
	}
	for _, dir := range loginShellPATHDirs() {
		add(dir)
	}
	for _, dir := range wellKnownInstallDirs() {
		add(dir)
	}
	return out
}

// Seams overridable in tests; production reads the real login shell and home.
var (
	readLoginShellPATH = defaultLoginShellPATH
	userHomeDir        = os.UserHomeDir
)

var (
	loginShellMu       sync.Mutex
	loginShellResolved bool
	loginShellCache    []string
)

// loginShellPATHDirs returns the directories from the user's login-shell PATH.
// The lookup spawns a shell, so the result is cached for the process lifetime.
func loginShellPATHDirs() []string {
	loginShellMu.Lock()
	defer loginShellMu.Unlock()
	if !loginShellResolved {
		loginShellCache = filepath.SplitList(readLoginShellPATH())
		loginShellResolved = true
	}
	return loginShellCache
}

func resetLoginShellCacheForTest() {
	loginShellMu.Lock()
	loginShellResolved = false
	loginShellCache = nil
	loginShellMu.Unlock()
}

// defaultLoginShellPATH asks the user's login shell for its PATH. It returns
// "" on Windows, when $SHELL is unset, or on any failure/timeout — callers
// then fall back to process PATH and well-known dirs.
func defaultLoginShellPATH() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), loginShellPATHTimeout)
	defer cancel()
	// -l sources login profiles (where PATH is typically extended); -c runs a
	// non-interactive command so the shell can never block on a prompt. The
	// markers fence $PATH off from any banner output the profile may print.
	script := "printf '%s' \"" + loginPATHMarkerStart + "${PATH}" + loginPATHMarkerEnd + "\""
	cmd := exec.CommandContext(ctx, shell, "-lc", script)
	cmd.Stdin = nil
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return ""
	}
	out := buf.String()
	start := strings.Index(out, loginPATHMarkerStart)
	end := strings.Index(out, loginPATHMarkerEnd)
	if start < 0 || end < 0 || end < start {
		return ""
	}
	return out[start+len(loginPATHMarkerStart) : end]
}

// wellKnownInstallDirs lists install locations commonly missing from a
// GUI/launchd PATH. Non-existent entries are harmless — the candidate scan
// stats each path and skips misses.
func wellKnownInstallDirs() []string {
	if runtime.GOOS == "windows" {
		return windowsWellKnownDirs()
	}
	dirs := []string{
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
	}
	home, err := userHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return dirs
	}
	dirs = append(
		dirs,
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, "bin"),
		filepath.Join(home, ".npm-global", "bin"),
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".deno", "bin"),
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, ".volta", "bin"),
		filepath.Join(home, ".asdf", "shims"),
		filepath.Join(home, ".cursor", "bin"),
		filepath.Join(home, ".claude", "bin"),
	)
	dirs = append(dirs, nodeVersionManagerBins(home)...)
	return dirs
}

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
