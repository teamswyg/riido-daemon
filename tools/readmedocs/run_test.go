package main

import "testing"

func TestRunAgainstRepositoryEvidence(t *testing.T) {
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderedDocsCoverReadmeHandoffPages(t *testing.T) {
	m, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(renderedDocs(m)); got != len(m.Fragments) {
		t.Fatalf("rendered doc count = %d, want %d", got, len(m.Fragments))
	}
}
