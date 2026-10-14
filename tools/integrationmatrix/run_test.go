package main

import "testing"

func TestRunAgainstRepositoryEvidence(t *testing.T) {
	if err := run(options{Repo: "../..", Manifest: defaultManifest}); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratedDocsCoverIntegrationSurface(t *testing.T) {
	m, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(renderedDocs(m)) != 6 {
		t.Fatalf("rendered docs = %d", len(renderedDocs(m)))
	}
}
