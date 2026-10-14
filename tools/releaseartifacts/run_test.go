package main

import "testing"

func TestRunAgainstRepositoryEvidence(t *testing.T) {
	if err := run(options{Repo: "../..", Manifest: defaultManifest}); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratedDocsCoverReleaseSurface(t *testing.T) {
	m, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(renderedDocs(m)) != 5 {
		t.Fatalf("rendered docs = %d", len(renderedDocs(m)))
	}
}
