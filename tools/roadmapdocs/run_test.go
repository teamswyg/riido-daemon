package main

import "testing"

func TestRunAgainstRepositoryEvidence(t *testing.T) {
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestManifestCarriesOpenQuestions(t *testing.T) {
	m, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Questions) != 7 {
		t.Fatalf("question count = %d", len(m.Questions))
	}
}
