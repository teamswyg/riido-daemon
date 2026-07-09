package main

import (
	"path/filepath"
	"testing"
)

func TestRepoPathKeepsAbsoluteAndJoinsRelative(t *testing.T) {
	repo := t.TempDir()
	abs := filepath.Join(repo, "already.txt")
	if got := repoPath(repo, abs); got != abs {
		t.Fatalf("absolute path changed: %q", got)
	}
	want := filepath.Join(repo, "docs", "validation.md")
	if got := repoPath(repo, "docs/validation.md"); got != want {
		t.Fatalf("relative path = %q, want %q", got, want)
	}
}

func TestChecksListFormatsMarkdownCodeList(t *testing.T) {
	if got := checksList(nil); got != "" {
		t.Fatalf("empty checks = %q", got)
	}
	if got := checksList([]string{"a", "b"}); got != "`a`, `b`" {
		t.Fatalf("checks markdown = %q", got)
	}
}

func TestValidateHeaderAcceptsCompleteManifestAndRejectsGaps(t *testing.T) {
	complete := Manifest{
		SchemaVersion:    "riido-validation-evidence.v1",
		ID:               "validation",
		Title:            "Validation",
		GeneratedDoc:     "docs/validation.md",
		Workflow:         ".github/workflows/validation.yml",
		EvidenceArtifact: "tmp/evidence.json",
		Purpose:          "prove validation evidence",
		Facts:            []Fact{{Name: "fact"}},
		Boundaries:       []Boundary{{Name: "boundary"}},
		SourceChecks:     []SourceCheck{{Name: "source"}},
	}
	if problems := validateHeader(complete); len(problems) != 0 {
		t.Fatalf("complete manifest rejected: %#v", problems)
	}

	missing := complete
	missing.SchemaVersion = "old"
	missing.Title = ""
	missing.Facts = nil
	problems := validateHeader(missing)
	if len(problems) != 3 {
		t.Fatalf("expected schema, required field, and collection errors: %#v", problems)
	}
}
