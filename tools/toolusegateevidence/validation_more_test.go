package main

import (
	"path/filepath"
	"testing"
)

func TestValidateSourcesRefsAndEvidence(t *testing.T) {
	dir := t.TempDir()
	if err := writeText(filepath.Join(dir, "source.go"), "needle"); err != nil {
		t.Fatal(err)
	}
	m := Manifest{
		SchemaVersion: "x",
		ID:            "tool-use",
		Title:         "Tool Use",
		GeneratedDoc:  "doc.md",
		Workflow:      "workflow.yml",
		ImplementedAction: []Action{{
			Name: "implemented", SourceChecks: []string{"missing"},
		}},
		ReservedActions: []Action{{Name: "reserved"}},
		Facts:           []Fact{{Name: "fact", SourceChecks: []string{"known"}}},
		SourceChecks: []SourceCheck{
			{Name: "known", File: "source.go", Contains: "needle"},
			{Name: "bad", File: "source.go", Contains: ""},
			{Name: "absent", File: "source.go", Contains: "absent"},
		},
		Assertions: []string{"tool use gates are explicit"},
	}
	problems, sources, absent := validate(dir, m)
	for _, want := range []string{
		"unknown source check missing",
		"missing source evidence bad",
		"missing source evidence absent",
	} {
		assertToolUseProblem(t, problems, want)
	}
	if len(sources) != 3 || len(absent) != 0 {
		t.Fatalf("unexpected evidence sources=%#v absent=%#v", sources, absent)
	}
	ev := buildEvidence(m, problems[:1], sources[:1], absent)
	if ev.ID != "tool-use" || len(ev.Problems) != 1 || len(ev.SourceChecks) != 1 {
		t.Fatalf("unexpected evidence %#v", ev)
	}
}

func TestValidateRefsStopsOnEmptySourceName(t *testing.T) {
	problems := validateRefs(Manifest{SourceChecks: []SourceCheck{{Name: ""}}})
	assertToolUseProblem(t, problems, "source check with empty name")
}
