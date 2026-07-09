package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanRepoPathRejectsAbsoluteAndParentTraversal(t *testing.T) {
	repo := t.TempDir()
	got, err := cleanRepoPath(repo, "docs/compatibility.json")
	if err != nil || got != filepath.Join(repo, "docs", "compatibility.json") {
		t.Fatalf("clean path = %q err=%v", got, err)
	}
	for _, rel := range []string{"/tmp/outside", "../outside", "docs/../outside"} {
		if _, err := cleanRepoPath(repo, rel); err == nil {
			t.Fatalf("expected unsafe path rejection for %q", rel)
		}
	}
}

func TestValidateRequiredAcceptsCompleteManifestAndReportsGaps(t *testing.T) {
	complete := compatibilityManifestFixture()
	if problems := validateRequired(complete); len(problems) != 0 {
		t.Fatalf("complete manifest rejected: %#v", problems)
	}
	missing := complete
	missing.Title = ""
	missing.Inputs = nil
	problems := validateRequired(missing)
	if len(problems) != 2 {
		t.Fatalf("expected required field and collection problems: %#v", problems)
	}
}

func TestValidateReferencesRejectsDuplicateAndUnknownRefs(t *testing.T) {
	dup := compatibilityManifestFixture()
	dup.SourceChecks = append(dup.SourceChecks, SourceCheck{Name: "source"})
	if problems := validateReferences(dup); len(problems) != 1 ||
		!strings.Contains(problems[0].Message, "duplicate source check") {
		t.Fatalf("duplicate source check not reported: %#v", problems)
	}
	problems := validateRefSet("input", "client", []string{"missing"},
		map[string]bool{"source": true})
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "unknown") {
		t.Fatalf("unknown source check not reported: %#v", problems)
	}
	if problems := validateRefSet("input", "client", nil, map[string]bool{}); len(problems) != 1 ||
		!strings.Contains(problems[0].Message, "has no source checks") {
		t.Fatalf("empty source refs not reported: %#v", problems)
	}
}

func compatibilityManifestFixture() Manifest {
	return Manifest{
		SchemaVersion:    "riido-compatibility-gate.v1",
		ID:               "compatibility",
		Title:            "Compatibility",
		GeneratedDoc:     "docs/compatibility.md",
		Workflow:         ".github/workflows/compatibility.yml",
		EvidenceArtifact: "tmp/compatibility.json",
		Purpose:          "prove compatibility",
		Inputs:           []GateInput{{Name: "client", SourceChecks: []string{"source"}}},
		GateOrder:        []GateStep{{Step: "verify", SourceChecks: []string{"source"}}},
		SourceChecks:     []SourceCheck{{Name: "source", File: "internal/source.go"}},
	}
}
