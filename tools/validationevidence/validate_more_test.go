package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildEvidenceMarksProblemsFailed(t *testing.T) {
	evidence := buildEvidence(
		Manifest{ID: "validation"},
		[]problem{{"missing source"}},
		[]SourceCheckResult{{Name: "src", Pass: false}},
		[]AbsentCheck{{Name: "secret", Pass: true}},
	)
	if evidence.Status != "failed" {
		t.Fatalf("expected failed status, got %s", evidence.Status)
	}
	if evidence.SchemaVersion != "riido-validation-evidence-result.v1" {
		t.Fatalf("schema=%q", evidence.SchemaVersion)
	}
	if len(evidence.Problems) != 1 || evidence.Problems[0] != "missing source" {
		t.Fatalf("unexpected problems: %#v", evidence.Problems)
	}
}

func TestValidateReportsReferenceAndAbsentDrift(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "src.go"), "secret token")
	m := Manifest{
		SchemaVersion:    "riido-validation-evidence.v1",
		ID:               "x",
		Title:            "X",
		GeneratedDoc:     "validation.md",
		Workflow:         ".github/workflows/x.yml",
		EvidenceArtifact: "e",
		Purpose:          "p",
		Facts:            []Fact{{Name: "fact", Summary: "s", SourceChecks: []string{"missing"}}},
		Boundaries:       []Boundary{{Name: "boundary", Owner: "owner", Summary: "s"}},
		SourceChecks:     []SourceCheck{{Name: "src", File: "src.go", Contains: "secret"}},
		AbsentSurfaces: []AbsentSurface{{
			Name:   "no-secrets",
			Scope:  []string{"src.go"},
			Tokens: []string{"secret"},
		}},
	}
	problems, sources, absent := validate(dir, m)
	if len(sources) != 1 || !sources[0].Pass {
		t.Fatalf("expected source check pass, got %#v", sources)
	}
	if len(absent) != 1 || absent[0].Pass || absent[0].Hits[0] != "src.go:secret" {
		t.Fatalf("expected absent hit, got %#v", absent)
	}
	if !hasProblem(problems, "unknown source check: missing") {
		t.Fatalf("missing source ref problem: %#v", problems)
	}
	if !hasProblem(problems, "absent surface found: no-secrets") {
		t.Fatalf("missing absent surface problem: %#v", problems)
	}
}

func hasProblem(problems []problem, fragment string) bool {
	for _, item := range problems {
		if strings.Contains(item.Message, fragment) {
			return true
		}
	}
	return false
}
