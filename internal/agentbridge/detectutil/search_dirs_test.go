package detectutil

import (
	"path/filepath"
	"testing"
)

func TestProductionSearchDirsIncludesLoginShellAndWellKnown(t *testing.T) {
	home := t.TempDir()
	loginDir := t.TempDir()
	restoreSearchDirSeams(t, home, loginDir)

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

func restoreSearchDirSeams(t *testing.T, home, loginDir string) {
	t.Helper()
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
}
