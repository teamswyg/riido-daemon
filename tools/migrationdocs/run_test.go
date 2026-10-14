package main

import (
	"strings"
	"testing"
)

func TestRunAgainstRepositoryEvidence(t *testing.T) {
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderedDocsCoverCLIMigrationPages(t *testing.T) {
	m, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(renderedDocs(m)) != m.ExpectedPageCount {
		t.Fatalf("rendered doc count = %d", len(renderedDocs(m)))
	}
}

func TestValidatePagesUsesManifestCount(t *testing.T) {
	problems := validatePages(nil, 2)
	if len(problems) != 1 || !strings.Contains(problems[0], "expected 2") {
		t.Fatalf("problems = %#v", problems)
	}
}
