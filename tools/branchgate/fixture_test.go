package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixtureRepo(t *testing.T, repo string) {
	t.Helper()
	writeFile(t, repo, defaultManifest, fixtureManifestSource, 0o644)
	writeFile(t, repo, ".github/workflows/branch.yml", "run: scripts/verify-branch.sh\n", 0o644)
	writeFile(t, repo, ".github/workflows/branch-evidence.yml", "run: go run ./tools/branchgate\n", 0o644)
}

func writeFile(t *testing.T, repo, rel, body string, perm os.FileMode) {
	t.Helper()
	path := filepath.Join(repo, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), perm); err != nil {
		t.Fatal(err)
	}
}
