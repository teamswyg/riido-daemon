package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteJSONAtomicCreatesParentAndFormatsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")

	if err := WriteJSONAtomic(path, map[string]string{"name": "riido"}); err != nil {
		t.Fatalf("WriteJSONAtomic() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	want := "{\n  \"name\": \"riido\"\n}\n"
	if string(content) != want {
		t.Fatalf("content = %q, want %q", string(content), want)
	}
}

func TestWriteFileAtomicReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.txt")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := WriteFileAtomic(path, []byte("new"), ""); err != nil {
		t.Fatalf("WriteFileAtomic() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("content = %q, want %q", string(content), "new")
	}
}
