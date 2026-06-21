package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePathAllowsMatchingGlob(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docs", "x.riido.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if problems := validatePath(dir, "owner", "docs/*.riido.json"); len(problems) != 0 {
		t.Fatalf("problems = %#v", problems)
	}
}

func TestValidatePathRejectsUnmatchedGlob(t *testing.T) {
	problems := validatePath(t.TempDir(), "owner", "docs/*.riido.json")
	if !containsLoopProblem(problems, "missing artifact glob") {
		t.Fatalf("problems = %#v", problems)
	}
}
