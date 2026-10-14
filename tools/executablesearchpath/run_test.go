package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAgainstRepository(t *testing.T) {
	repo := testRepoRoot(t)
	out := filepath.Join(t.TempDir(), "evidence.json")
	err := run(options{Repo: repo, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("evidence not written: %v", err)
	}
}

func testRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			t.Fatal("go.mod not found")
		}
		dir = next
	}
}
