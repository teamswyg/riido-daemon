package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateManifestReportsQuestionAndSourceProblems(t *testing.T) {
	dir := t.TempDir()
	mustRoadmapWrite(t, filepath.Join(dir, ".github", "workflows", "qa.yml"), "workflow")
	mustRoadmapWrite(t, filepath.Join(dir, "src.md"), "anchor")
	m := manifest{
		SchemaVersion:    manifestSchema,
		ID:               "roadmap",
		Title:            "Roadmap",
		GeneratedDoc:     "roadmap.md",
		Workflow:         ".github/workflows/qa.yml",
		EvidenceArtifact: "evidence.json",
		Questions: []question{
			{ID: "Q-1", Area: "A", Question: "?", CurrentHandling: "manual"},
			{ID: "bad", Area: "A", Question: "?", CurrentHandling: "manual"},
			{ID: "Q-1", Area: "B", Question: "?", CurrentHandling: "manual"},
		},
		SourceChecks: []sourceCheck{{Name: "src", File: "src.md", Contains: "missing"}},
		Assertions:   []string{"assert"},
	}
	problems, checks := validateManifest(dir, m)
	if len(checks) != 1 || checks[0].Passed {
		t.Fatalf("expected failed source check, got %#v", checks)
	}
	for _, want := range []string{"question id must start", "duplicate question id Q-1", "source check"} {
		if !containsProblem(problems, want) {
			t.Fatalf("missing %q in %#v", want, problems)
		}
	}
}

func containsProblem(problems []string, fragment string) bool {
	for _, item := range problems {
		if strings.Contains(item, fragment) {
			return true
		}
	}
	return false
}

func mustRoadmapWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
