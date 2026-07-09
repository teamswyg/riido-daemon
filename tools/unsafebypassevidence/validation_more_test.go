package main

import (
	"path/filepath"
	"testing"
)

func TestValidateReportsRequiredRefsSourcesAndPolicyDrift(t *testing.T) {
	dir := t.TempDir()
	if err := writeText(filepath.Join(dir, "source.go"), "needle"); err != nil {
		t.Fatal(err)
	}
	m := Manifest{
		SchemaVersion: "x",
		ID:            "unsafe",
		Title:         "Unsafe",
		GeneratedDoc:  "doc.md",
		Workflow:      "workflow.yml",
		Surfaces: []Surface{{
			Provider: "Codex", Surface: "unknown",
			Flag: "--missing=true", SourceChecks: []string{"missing"},
		}},
		SourceChecks: []SourceCheck{
			{Name: "dup", File: "source.go", Contains: "needle"},
			{Name: "dup", File: "source.go", Contains: "needle"},
			{Name: "bad", File: "source.go", Contains: ""},
			{Name: "absent", File: "source.go", Contains: "absent"},
			{Name: "missing-file", File: "missing.go", Contains: "needle"},
		},
		Assertions: []string{"host denies unsafe bypass"},
	}
	problems, sources, policies, codexArgs := validate(dir, m)
	for _, want := range []string{
		"duplicate source check dup",
		"missing source evidence bad",
		"missing source evidence absent",
		"missing-file",
		"unsafe bypass policy drift for unknown",
		"codex unsafe arg missing: --missing",
	} {
		assertUnsafeProblem(t, problems, want)
	}
	if len(sources) != 5 || len(policies) != 1 || len(codexArgs) != 1 {
		t.Fatalf("unexpected evidence sources=%#v policies=%#v codex=%#v", sources, policies, codexArgs)
	}
	refProblems := validateRefs(Manifest{
		Surfaces:     []Surface{{SourceChecks: []string{"missing"}}},
		SourceChecks: []SourceCheck{{Name: "known"}},
	})
	assertUnsafeProblem(t, refProblems, "unknown source check missing")
}

func TestBuildEvidenceCopiesProblemsAndSemanticRows(t *testing.T) {
	ev := buildEvidence(
		Manifest{ID: "unsafe", EvidenceArtifact: "out.json", Assertions: []string{"assert"}},
		[]problem{{Message: "bad"}},
		[]SourceEvidence{{Name: "source", File: "source.go", OK: true}},
		[]PolicyEvidence{{Surface: "codex", OK: true}},
		[]CodexArgEvidence{{Arg: "--dangerously-bypass-approvals-and-sandbox", OK: true}},
	)
	if ev.ID != "unsafe" || ev.Artifact != "out.json" || len(ev.Problems) != 1 {
		t.Fatalf("unexpected evidence header %#v", ev)
	}
	if len(ev.SourceChecks) != 1 || len(ev.PolicyChecks) != 1 ||
		len(ev.CodexArgChecks) != 1 || len(ev.Assertions) != 1 {
		t.Fatalf("semantic evidence not copied %#v", ev)
	}
}
