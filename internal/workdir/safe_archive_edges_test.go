package workdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeJoinRejectsUnsafePaths(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"", "/absolute", "..", "../escape", "nested/../../escape"} {
		if got, err := safeJoin(root, rel); err == nil {
			t.Fatalf("safeJoin(%q) = %q, want error", rel, got)
		}
	}
	got, err := safeJoin(root, "nested/../file.txt")
	if err != nil {
		t.Fatalf("safeJoin should allow normalized in-root path: %v", err)
	}
	if got != filepath.Join(root, "file.txt") {
		t.Fatalf("safeJoin normalized path = %q", got)
	}
}

func TestWriteFileUnderDefaultsModeAndRejectsEscape(t *testing.T) {
	root := t.TempDir()
	if err := writeFileUnder(root, "nested/file.txt", []byte("ok"), 0); err != nil {
		t.Fatalf("writeFileUnder: %v", err)
	}
	path := filepath.Join(root, "nested", "file.txt")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat written file: %v", err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("mode=%#o, want 0644", info.Mode().Perm())
	}
	if bytes, err := os.ReadFile(path); err != nil || string(bytes) != "ok" {
		t.Fatalf("written content=%q err=%v", bytes, err)
	}
	if err := writeFileUnder(root, "../escape.txt", []byte("bad"), 0); err == nil {
		t.Fatalf("expected escaping write to fail")
	}
}

func TestArchiveRejectsInvalidWorkspaceOrStatus(t *testing.T) {
	adapter := NewFSAdapter(t.TempDir())
	valid := Workspace{Root: filepath.Join(t.TempDir(), "run"), Workdir: t.TempDir()}
	for name, ws := range map[string]Workspace{
		"missing root":    {Workdir: valid.Workdir},
		"missing workdir": {Root: valid.Root},
	} {
		if _, err := adapter.Archive(ws, ArchiveRequest{ResultStatus: "completed"}); err == nil {
			t.Fatalf("%s: expected workspace validation error", name)
		}
	}
	_, err := adapter.Archive(valid, ArchiveRequest{ResultStatus: "  "})
	if err == nil || !strings.Contains(err.Error(), "result status") {
		t.Fatalf("expected result status error, got %v", err)
	}
}
