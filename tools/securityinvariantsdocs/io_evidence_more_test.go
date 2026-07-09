package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildEvidenceAndProblemStatus(t *testing.T) {
	m := manifest{
		ID:               "security-invariants",
		GeneratedDoc:     "root.md",
		EvidenceArtifact: "artifact",
		Assertions:       []string{"one", "two"},
		DetailPages:      []detailLink{{ID: "detail", Path: "detail.md"}},
		Details:          []detailDoc{{ID: "detail"}},
	}
	checks := []sourceCheckResult{{Name: "source", File: "file", Passed: true}}
	ok := buildEvidence(m, checks, nil)
	if ok.Status != "verified" || ok.DetailCount != 1 || ok.AssertionCount != 2 {
		t.Fatalf("verified evidence = %+v", ok)
	}
	failed := buildEvidence(m, checks, []string{"missing"})
	if failed.Status != "failed" || problemError(failed.ProblemSummaries).Error() != "missing" {
		t.Fatalf("failed evidence = %+v", failed)
	}
}

func TestReadJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.json")
	if err := os.WriteFile(valid, []byte(`{"schema_version":"v","id":"id"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var detail detailDoc
	if err := readJSON(valid, &detail); err != nil {
		t.Fatal(err)
	}
	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"schema_version":"v","id":"id","extra":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(unknown, &detail); err == nil {
		t.Fatal("expected unknown field to fail")
	}
	trailing := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{"schema_version":"v","id":"id"} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(trailing, &detail); err == nil {
		t.Fatal("expected trailing JSON value to fail")
	}
}

func TestWriteTextAndJSONCreateParents(t *testing.T) {
	dir := t.TempDir()
	text := filepath.Join(dir, "nested", "doc.md")
	if err := writeText(text, "hello"); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(text); err != nil || string(got) != "hello" {
		t.Fatalf("text = %q err=%v", got, err)
	}
	jsonPath := filepath.Join(dir, "nested", "evidence.json")
	if err := writeJSON(jsonPath, map[string]string{"status": "verified"}); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(jsonPath); err != nil || string(got) != "{\n  \"status\": \"verified\"\n}\n" {
		t.Fatalf("json = %q err=%v", got, err)
	}
}
