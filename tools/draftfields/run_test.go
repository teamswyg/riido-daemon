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

func TestRunDetectsForbiddenFieldLeak(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, fixtureSourceFiles[0].path, "package supervisor\nvar _ = EventID\n")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest})
	if err == nil {
		t.Fatal("expected forbidden field leak")
	}
}
