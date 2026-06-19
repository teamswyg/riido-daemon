package main

import "testing"

func TestRunAgainstRepositoryEvidence(t *testing.T) {
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderedDocsCoverProviderRuntimeResponsibilitySurfaces(t *testing.T) {
	m, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(renderedDocs(m)) != 9 {
		t.Fatalf("rendered doc count = %d", len(renderedDocs(m)))
	}
}
