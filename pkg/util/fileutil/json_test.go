package fileutil

import (
	"os"
	"path/filepath"
	"strings"
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

func TestWriteJSONAtomicReturnsMarshalError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	err := WriteJSONAtomic(path, map[string]chan int{"bad": make(chan int)})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected marshal error, got %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("target should not be created, stat=%v", statErr)
	}
}

func TestWriteJSONAtomicFailsWhenParentIsFile(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "state-parent")
	if err := os.WriteFile(parent, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := WriteJSONAtomic(filepath.Join(parent, "state.json"), map[string]string{"x": "y"})
	if err == nil || !strings.Contains(err.Error(), "create parent directory") {
		t.Fatalf("expected parent error, got %v", err)
	}
}

func TestWriteFileAtomicRequiresExistingParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "state.txt")
	err := WriteFileAtomic(path, []byte("new"), "")
	if err == nil {
		t.Fatal("expected temp file creation error")
	}
}
