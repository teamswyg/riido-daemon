package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepoPathKeepsAbsoluteAndRootsRelativePaths(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "manifest.json")
	if got := repoPath("/repo", abs); got != abs {
		t.Fatalf("absolute path changed: %q", got)
	}
	if got := repoPath("/repo", "docs/privacy.json"); got != filepath.Join("/repo", "docs/privacy.json") {
		t.Fatalf("relative path not rooted: %q", got)
	}
}

func TestWriteTextReportsParentCreationError(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "not-dir"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := writeText(filepath.Join(root, "not-dir", "doc.md"), "doc")
	if err == nil {
		t.Fatal("expected parent creation error")
	}
}

func TestWriteJSONReportsMarshalAndWriteErrors(t *testing.T) {
	root := t.TempDir()
	err := writeJSON(filepath.Join(root, "evidence.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected marshal error, got %v", err)
	}
	err = writeJSON(filepath.Join(root, "missing", "evidence.json"), map[string]string{"ok": "true"})
	if err == nil || !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("expected write error, got %v", err)
	}
}
