package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateAssertionsAndReportSummaries(t *testing.T) {
	item := requiredEvidence{
		ID: "loop",
		Assertions: []assertion{
			{Field: "status", Equals: "verified"},
			{Field: "problems", Empty: true},
			{Field: "missing", Equals: "value"},
		},
	}
	checks, problems := evaluate(item, map[string]any{
		"status":   "failed",
		"problems": []any{"open"},
	})
	for _, want := range []string{
		"loop status = failed, want verified",
		"loop problems is not empty",
		"loop missing missing",
	} {
		assertSelfProblem(t, problems, want)
	}
	if len(checks) != 3 {
		t.Fatalf("checks=%#v", checks)
	}
}

func TestBuildReportMarksDependentClosedLoopsOpen(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeFixtureManifest(t, root)
	m, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	evidenceDir := filepath.Join(root, "out")
	mustMkdir(t, evidenceDir)
	mustWrite(t, filepath.Join(evidenceDir, "loop.json"), `{"status":"failed","problem_count":1}`)
	report := buildReport(evidenceDir, m)
	if report.Status != statusFailed || report.ClosedVerified != 0 || report.ProblemCount == 0 {
		t.Fatalf("unexpected report %#v", report)
	}
	assertSelfProblem(t, report.Problems, "bug is open because loop failed")
}

func assertSelfProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
