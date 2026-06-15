package detectutil

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestResolveExecutableEnvOverrideWins(t *testing.T) {
	override := filepath.Join(t.TempDir(), "fake-claude")
	if err := os.WriteFile(override, []byte("#!/bin/sh\necho hi\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, ok := ResolveExecutable("claude", override)
	if !ok || got != override {
		t.Fatalf("override should win: %q ok=%v", got, ok)
	}
}

// An override set to a non-existent path must fail closed, NOT fall
// back to PATH lookup. Otherwise a misconfigured RIIDO_*_PATH would
// silently run a different binary than the operator chose.
func TestResolveExecutableMissingOverrideFailsClosed(t *testing.T) {
	got, ok := ResolveExecutable("sh", "/definitely/not/real-xyz")
	if ok {
		t.Fatalf("override pointing nowhere must NOT fall back to PATH, got %q", got)
	}
}

func TestResolveExecutablePathFallback(t *testing.T) {
	got, ok := ResolveExecutable("sh", "")
	if !ok || got == "" {
		t.Fatalf("sh should be on PATH: %q ok=%v", got, ok)
	}
}

func TestResolveExecutableCandidatesPreservePathOrder(t *testing.T) {
	firstDir := t.TempDir()
	secondDir := t.TempDir()
	first := writeExecutable(t, filepath.Join(firstDir, "fake-tool"), "first")
	second := writeExecutable(t, filepath.Join(secondDir, "fake-tool"), "second")
	t.Setenv("PATH", firstDir+string(os.PathListSeparator)+secondDir)

	got := ResolveExecutableCandidates("fake-tool", "")
	if len(got) != 2 {
		t.Fatalf("candidate count: got %d (%v), want 2", len(got), got)
	}
	if got[0] != first || got[1] != second {
		t.Fatalf("candidate order: got %v, want [%s %s]", got, first, second)
	}
}

func TestResolveExecutableCandidatesOverrideIsOnlyCandidate(t *testing.T) {
	override := writeExecutable(t, filepath.Join(t.TempDir(), "fake-tool"), "override")
	pathDir := t.TempDir()
	_ = writeExecutable(t, filepath.Join(pathDir, "fake-tool"), "path")
	t.Setenv("PATH", pathDir)

	got := ResolveExecutableCandidates("fake-tool", override)
	if len(got) != 1 || got[0] != override {
		t.Fatalf("override must be the only candidate, got %v", got)
	}
}

func TestResolveExecutableCandidatesMissingOverrideFailsClosed(t *testing.T) {
	pathDir := t.TempDir()
	_ = writeExecutable(t, filepath.Join(pathDir, "fake-tool"), "path")
	t.Setenv("PATH", pathDir)

	got := ResolveExecutableCandidates("fake-tool", "/definitely/not/real-xyz")
	if len(got) != 0 {
		t.Fatalf("missing override must not fall back to PATH, got %v", got)
	}
}

func TestResolveExecutableMissing(t *testing.T) {
	_, ok := ResolveExecutable("definitely-not-a-real-binary-xyz", "")
	if ok {
		t.Fatal("expected not found")
	}
}

// A GUI/launchd-spawned daemon inherits a minimal process PATH that omits
// the directory where the provider CLI is installed. Resolution must still
// find the binary through the augmented search dirs (login-shell PATH and
// well-known install locations).
func TestResolveExecutableFindsToolOutsideProcessPATH(t *testing.T) {
	minimalDir := t.TempDir()
	toolDir := t.TempDir()
	tool := writeExecutable(t, filepath.Join(toolDir, "fake-claude"), "v1")
	t.Setenv("PATH", minimalDir)

	restore := augmentedSearchDirs
	augmentedSearchDirs = func() []string { return []string{minimalDir, toolDir} }
	t.Cleanup(func() { augmentedSearchDirs = restore })

	got, ok := ResolveExecutable("fake-claude", "")
	if !ok || got != tool {
		t.Fatalf("want %s found via augmented dirs, got %q ok=%v", tool, got, ok)
	}
}

func TestProductionSearchDirsIncludesLoginShellAndWellKnown(t *testing.T) {
	home := t.TempDir()
	loginDir := t.TempDir()

	restoreShell := readLoginShellPATH
	restoreHome := userHomeDir
	readLoginShellPATH = func() string { return loginDir }
	userHomeDir = func() (string, error) { return home, nil }
	resetLoginShellCacheForTest()
	t.Cleanup(func() {
		readLoginShellPATH = restoreShell
		userHomeDir = restoreHome
		resetLoginShellCacheForTest()
	})

	t.Setenv("PATH", "/usr/bin")
	dirs := productionSearchDirs()

	if len(dirs) == 0 || dirs[0] != "/usr/bin" {
		t.Fatalf("os PATH must come first, got %v", dirs)
	}
	if !containsDir(dirs, loginDir) {
		t.Fatalf("login-shell PATH dir %q missing from %v", loginDir, dirs)
	}
	if want := filepath.Join(home, ".local", "bin"); !containsDir(dirs, want) {
		t.Fatalf("well-known dir %q missing from %v", want, dirs)
	}
}

// The login-shell PATH lookup must be cached so we do not spawn a shell on
// every Detect call.
func TestLoginShellPATHDirsCached(t *testing.T) {
	calls := 0
	restore := readLoginShellPATH
	readLoginShellPATH = func() string {
		calls++
		return "/tmp/cached-bin"
	}
	resetLoginShellCacheForTest()
	t.Cleanup(func() {
		readLoginShellPATH = restore
		resetLoginShellCacheForTest()
	})

	_ = loginShellPATHDirs()
	_ = loginShellPATHDirs()
	if calls != 1 {
		t.Fatalf("login-shell PATH should be read once, got %d reads", calls)
	}
}

func containsDir(dirs []string, want string) bool {
	return slices.Contains(dirs, want)
}

func writeExecutable(t *testing.T, path, output string) string {
	t.Helper()
	script := "#!/bin/sh\necho '" + output + "'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
	return path
}

func TestVersionProbeEchoesOutput(t *testing.T) {
	out, ok := VersionProbe(context.Background(), "/bin/echo", "1.2.3")
	if !ok {
		t.Fatal("probe failed")
	}
	if out != "1.2.3" {
		t.Fatalf("output: %q", out)
	}
}

func TestVersionProbeMissingBinary(t *testing.T) {
	_, ok := VersionProbe(context.Background(), "", "--version")
	if ok {
		t.Fatal("empty exe should fail")
	}
	_, ok = VersionProbe(context.Background(), "/no/such/path", "--version")
	if ok {
		t.Fatal("missing binary should fail")
	}
}

func TestVersionProbeStrictReportsExitCodeAndOutput(t *testing.T) {
	res := VersionProbeStrict(context.Background(), "/bin/sh", "-c", "printf 'tool failed'; exit 7")
	if !res.OK {
		t.Fatal("strict probe should report command completion")
	}
	if res.ExitCode != 7 {
		t.Fatalf("exit code: %d", res.ExitCode)
	}
	if res.Output != "tool failed" {
		t.Fatalf("output: %q", res.Output)
	}
}

func TestVersionProbeStrictMissingBinary(t *testing.T) {
	res := VersionProbeStrict(context.Background(), "/no/such/path", "--version")
	if res.OK {
		t.Fatalf("missing binary should fail closed: %+v", res)
	}
}
