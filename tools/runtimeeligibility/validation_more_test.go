package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/scheduling"
)

func TestValidateReportsSourcesRefsGatesAndAbsentScans(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "source.go"), []byte("anchor forbidden"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := Manifest{
		SchemaVersion: "x",
		ID:            "runtime",
		Title:         "Runtime",
		GeneratedDoc:  "doc.md",
		Workflow:      "workflow.yml",
		Inputs:        []Input{{Name: "input", SourceChecks: []string{"missing"}}},
		Gates: []Gate{
			{Order: 2, Code: "PROVIDER_MISMATCH", SourceChecks: []string{"ok"}},
			{Order: 2, Code: "UNKNOWN_GATE", SourceChecks: []string{"ok"}},
		},
		SourceChecks: []SourceCheck{
			{Name: "ok", File: "source.go", Contains: "anchor"},
			{Name: "ok", File: "source.go", Contains: "anchor"},
			{Name: "absent", File: "source.go", Contains: "absent"},
		},
		AbsentScans: []AbsentScan{{
			Name: "forbidden runtime shortcut", Scope: []string{"source.go"}, Tokens: []string{"forbidden"},
		}},
	}
	problems, sources, gates, absent := validate(dir, m)
	for _, want := range []string{
		"input references unknown check missing",
		"duplicate source check ok",
		"source.go missing absent",
		"PROVIDER_MISMATCH order drift",
		"unknown gate code UNKNOWN_GATE",
		"source.go contains forbidden token forbidden",
	} {
		assertRuntimeProblem(t, problems, want)
	}
	if len(sources) != 2 || len(gates) != 1 || len(absent) != 1 {
		t.Fatalf("unexpected evidence sources=%#v gates=%#v absent=%#v", sources, gates, absent)
	}
}

func TestBuildEvidenceAndGateScenarios(t *testing.T) {
	problems := []problem{{Message: "bad"}}
	ev := buildEvidence(
		Manifest{ID: "runtime", Assertions: []string{"assert"}},
		problems,
		[]SourceEvidence{{Name: "source", File: "file.go"}},
		[]GateEvidence{{Order: 1, Code: "PROVIDER_MISMATCH", Seen: true}},
		[]AbsentEvidence{{Name: "absent"}},
	)
	if ev.ID != "runtime" || ev.ProblemCount != 1 || len(ev.Assertions) != 1 {
		t.Fatalf("unexpected evidence %#v", ev)
	}
	_, _, ok := gateScenario("UNKNOWN_GATE")
	if ok {
		t.Fatal("unknown gate scenario reported ok")
	}
	if hasReason(scheduling.Eligibility{}, "anything") {
		t.Fatal("empty eligibility reasons matched")
	}
}
