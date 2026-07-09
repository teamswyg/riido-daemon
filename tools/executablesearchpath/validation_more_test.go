package main

import (
	"path/filepath"
	"testing"
)

func TestValidateSourcesReportsAllFailures(t *testing.T) {
	dir := t.TempDir()
	if err := writeText(filepath.Join(dir, "ok.txt"), "needle"); err != nil {
		t.Fatal(err)
	}
	checks := []SourceCheck{
		{Name: "ok", File: "ok.txt", Contains: "needle"},
		{Name: "ok", File: "ok.txt", Contains: "needle"},
		{Name: "missing", File: "missing.txt", Contains: "needle"},
		{Name: "absent", File: "ok.txt", Contains: "absent"},
	}
	problems, evidence := validateSources(dir, checks)
	for _, want := range []string{
		"duplicate source check ok",
		"missing.txt",
		"ok.txt missing absent",
	} {
		assertExecPathProblem(t, problems, want)
	}
	if len(evidence) != 2 {
		t.Fatalf("expected two successful source entries, got %#v", evidence)
	}
}

func TestValidateRefsAndBehaviorsReportContractDrift(t *testing.T) {
	m := Manifest{
		SearchOrder: []SearchStep{{
			Name: "search", SourceChecks: []string{"unknown"}, Behavior: "unknown",
		}},
		Rules: []Rule{{Name: "rule"}},
		SourceChecks: []SourceCheck{{
			Name: "known", File: "file.go", Contains: "needle",
		}},
	}
	for _, want := range []string{
		"search references unknown check unknown",
		"rule has no source checks",
	} {
		assertExecPathProblem(t, validateRefs(m), want)
	}
	problems, evidence := validateBehaviors(m)
	assertExecPathProblem(t, problems, "unknown behavior unknown")
	if len(evidence) != 1 || evidence[0].OK {
		t.Fatalf("expected failed behavior evidence, got %#v", evidence)
	}
}

func TestBuildEvidenceCopiesSemanticFields(t *testing.T) {
	ev := buildEvidence(
		Manifest{ID: "exec", Assertions: []string{"PATH is frozen"}},
		[]problem{{Message: "bad"}},
		[]SourceEvidence{{Name: "source", File: "file.go"}},
		[]BehaviorEvidence{{Name: "path_order", OK: true}},
	)
	if ev.SchemaVersion == "" || ev.ID != "exec" || ev.ProblemCount != 1 {
		t.Fatalf("unexpected evidence header %#v", ev)
	}
	if len(ev.Sources) != 1 || len(ev.Behaviors) != 1 || len(ev.Assertions) != 1 {
		t.Fatalf("semantic evidence not copied %#v", ev)
	}
}
