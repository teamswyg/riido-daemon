package detectutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveExecutableWindowsNativeClaudeOutsidePATH(t *testing.T) {
	home := t.TempDir()
	exe := filepath.Join(home, ".local", "bin", "claude.exe")
	if err := os.MkdirAll(filepath.Dir(exe), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, []byte("fixture"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	original := augmentedSearchDirs
	augmentedSearchDirs = func() []string { return windowsHomeDirs(home) }
	t.Cleanup(func() { augmentedSearchDirs = original })
	got, ok := ResolveExecutable("claude.exe", "")
	if !ok || got != exe {
		t.Fatalf("native Windows Claude = %q, found = %v; want %q", got, ok, exe)
	}
}
