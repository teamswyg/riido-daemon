package detectutil

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

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

func envListValue(env []string, wantKey string) (string, bool) {
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(key, wantKey) {
			return value, true
		}
	}
	return "", false
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
