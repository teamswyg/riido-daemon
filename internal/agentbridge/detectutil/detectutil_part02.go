package detectutil

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func envListHasPATH(env []string) bool {
	for _, entry := range env {
		key, _, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(key, pathEnvKey()) {
			return true
		}
	}
	return false
}

func pathEnvKey() string {
	if runtime.GOOS == "windows" {
		return "Path"
	}
	return "PATH"
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
