package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildEvidenceAndProblemHelpers(t *testing.T) {
	m := manifest{
		ID:               "roadmap",
		GeneratedDoc:     "docs/open.md",
		EvidenceArtifact: "evidence.json",
		Questions:        []question{{ID: "Q-1"}},
		Assertions:       []string{"assertion"},
	}
	checks := []sourceCheckResult{{Name: "anchor", File: "README.md", Passed: true}}
	ev := buildEvidence(m, checks, []string{"missing artifact"})
	if ev.Status != "failed" || ev.QuestionCount != 1 || ev.AssertionCount != 1 {
		t.Fatalf("unexpected evidence=%+v", ev)
	}
	if ev.SourceChecks[0].Name != "anchor" || len(ev.ProblemSummaries) != 1 {
		t.Fatalf("missing checks/problems=%+v", ev)
	}
	if statusFor(nil) != "verified" {
		t.Fatal("empty problem set should verify")
	}
	if !strings.Contains(problemError([]string{"a", "b"}).Error(), "a\nb") {
		t.Fatal("problem error should preserve problem list")
	}
}

func TestRoadmapDocsIOWritesAndRejectsNonSemanticJSON(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "nested", "doc.md")
	if err := writeText(textPath, "hello"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(textPath); err != nil || string(data) != "hello" {
		t.Fatalf("text=%q err=%v", data, err)
	}
	jsonPath := filepath.Join(dir, "evidence.json")
	if err := writeJSON(jsonPath, map[string]string{"id": "roadmap"}); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(jsonPath); err != nil || !strings.HasSuffix(string(data), "\n") {
		t.Fatalf("json=%q err=%v", data, err)
	}
	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"unknown":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := readJSON(unknown, &m); err == nil {
		t.Fatal("unknown manifest fields must fail")
	}
	trailing := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(trailing, &m); err == nil {
		t.Fatal("trailing JSON values must fail")
	}
}
