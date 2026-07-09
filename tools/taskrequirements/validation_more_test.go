package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateReportsSourcesRefsAndSurfaceDrift(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "source.go"), []byte("anchor"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := Manifest{
		SchemaVersion: "x",
		ID:            "task",
		Title:         "Task Requirements",
		GeneratedDoc:  "doc.md",
		Workflow:      "workflow.yml",
		Surfaces: []Surface{
			{Name: "mcp", CapabilityFlag: "wrong", SchedulingConstant: "SurfaceMCP",
				CandidateField: "SupportsMCP", SourceChecks: []string{"ok"}},
			{Name: "mcp", CapabilityFlag: "SupportsMCP", SchedulingConstant: "SurfaceMCP",
				CandidateField: "SupportsMCP", SourceChecks: []string{"missing"}},
			{Name: "future", SourceChecks: []string{"ok"}},
		},
		Inputs: []Input{{Name: "input", SourceChecks: []string{}}},
		SourceChecks: []SourceCheck{
			{Name: "ok", File: "source.go", Contains: "anchor"},
			{Name: "ok", File: "source.go", Contains: "anchor"},
			{Name: "absent", File: "source.go", Contains: "absent"},
		},
		Assertions: []string{"required surfaces fail closed"},
	}
	problems, sources, surfaces := validate(dir, m)
	for _, want := range []string{
		"mcp references unknown check missing",
		"input has no source checks",
		"duplicate source check ok",
		"source.go missing absent",
		"mcp capability flag drift",
		"duplicate surface mcp",
		"unknown manifest surface future",
		"future did not produce missing required surface",
		"future was not eligible when candidate flag was true",
	} {
		assertTaskReqProblem(t, problems, want)
	}
	if len(sources) != 2 || len(surfaces) != 3 {
		t.Fatalf("unexpected evidence sources=%#v surfaces=%#v", sources, surfaces)
	}
}

func TestBuildEvidenceAndUnknownSurfaceFailClosed(t *testing.T) {
	problems := validateUnknownSurfaceFailsClosed()
	if len(problems) != 0 {
		t.Fatalf("unknown surface should fail closed cleanly: %#v", problems)
	}
	ev := buildEvidence(
		Manifest{ID: "task", Assertions: []string{"assert"}},
		[]problem{{Message: "bad"}},
		[]SourceEvidence{{Name: "source", File: "file.go"}},
		[]SurfaceEvidence{{Name: "mcp", MissingCode: "MISSING_REQUIRED_SURFACE"}},
	)
	if ev.ID != "task" || ev.ProblemCount != 1 || len(ev.Surfaces) != 1 {
		t.Fatalf("unexpected evidence %#v", ev)
	}
}
