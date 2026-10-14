package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixtureFile(t *testing.T, repo, path string) {
	t.Helper()
	writeFixtureFileWithData(t, repo, path, []byte(path+"\n"))
}

func writeFixtureFileWithData(t *testing.T, repo, path string, data []byte) {
	t.Helper()
	full := filepath.Join(repo, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
