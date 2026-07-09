package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanRepoPathCoversSafeAndUnsafePaths(t *testing.T) {
	got, err := cleanRepoPath("/repo", "docs/process.json")
	if err != nil {
		t.Fatalf("safe path rejected: %v", err)
	}
	if got != filepath.Join("/repo", "docs", "process.json") {
		t.Fatalf("path=%q", got)
	}
	for _, rel := range []string{"/tmp/process.json", "../process.json", "docs/../secret.json"} {
		if _, err := cleanRepoPath("/repo", rel); err == nil {
			t.Fatalf("unsafe path %q accepted", rel)
		}
	}
}

func TestReadSourceAndLoadManifestRejectUnsafePaths(t *testing.T) {
	root := t.TempDir()
	if _, err := readSource(root, "../source.go"); err == nil {
		t.Fatal("expected unsafe source path to fail")
	}
	if _, err := loadManifest(root, "../manifest.json"); err == nil {
		t.Fatal("expected unsafe manifest path to fail")
	}
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("marker"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readSource(root, "source.go")
	if err != nil || got != "marker" {
		t.Fatalf("source=%q err=%v", got, err)
	}
}

func TestWriteJSONReportsMarshalAndWriteErrors(t *testing.T) {
	root := t.TempDir()
	err := writeJSON(filepath.Join(root, "bad.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected marshal error, got %v", err)
	}
	err = writeJSON(root, Manifest{ID: "x"})
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestProblemMessagesPreservesOrder(t *testing.T) {
	got := problemMessages([]problem{{"first"}, {"second"}})
	if strings.Join(got, ",") != "first,second" {
		t.Fatalf("messages=%v", got)
	}
}
