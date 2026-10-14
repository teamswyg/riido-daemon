package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesVerifiedEvidence(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	out := filepath.Join(repo, "out.json")
	err := run(t.Context(), options{
		Repo: repo, Manifest: defaultManifest,
		WriteDoc: true, CheckDoc: true, EvidenceOut: out,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunDetectsForbiddenSurface(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "internal/handler.go", "package internal\nconst endpoint = \"thread-progress\"\n")
	if err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest}); err == nil {
		t.Fatal("expected forbidden thread-progress surface")
	}
}
