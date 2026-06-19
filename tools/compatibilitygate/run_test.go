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

func TestRunDetectsSourceDrift(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "internal/source.go", "package internal\n")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest})
	if err == nil {
		t.Fatal("expected source drift")
	}
}
