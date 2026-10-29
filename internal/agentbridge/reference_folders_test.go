package agentbridge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckReferenceDirectories(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "자료 with spaces")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := CheckReferenceDirectories([]string{dir}); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{file, filepath.Join(root, "deleted"), "relative"} {
		if err := CheckReferenceDirectories([]string{path}); err == nil || !strings.Contains(err.Error(), path) {
			t.Fatalf("missing path diagnostic for %q: %v", path, err)
		}
	}
	if err := CheckReferenceDirectories(nil); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0700)
	if _, err := os.ReadDir(dir); err != nil {
		if err := CheckReferenceDirectories([]string{dir}); err == nil || !strings.Contains(err.Error(), dir) {
			t.Fatalf("unreadable folder must fail with its path: %v", err)
		}
	}
}
