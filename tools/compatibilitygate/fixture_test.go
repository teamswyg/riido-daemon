package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixtureRepo(t *testing.T, repo string) {
	t.Helper()
	writeFile(t, repo, defaultManifest, fixtureManifestSource)
	writeFile(t, repo, "internal/source.go", "package internal\nconst _ = \"needle\"\n")
}

func writeFile(t *testing.T, repo, rel, body string) {
	t.Helper()
	path := filepath.Join(repo, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
