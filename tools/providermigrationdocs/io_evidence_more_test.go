package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildEvidenceAndProblemStatus(t *testing.T) {
	m := manifest{
		ID:               "provider-migration",
		GeneratedDoc:     "root.md",
		EvidenceArtifact: "artifact",
		Assertions:       []string{"one"},
		Pages: []page{
			{GeneratedDoc: "a.md", ProviderID: "claude", Artifacts: []string{"a", "b"}},
			{GeneratedDoc: "b.md", Artifacts: []string{"c"}},
		},
	}
	ok := buildEvidence(m, nil)
	if ok.Status != "verified" || ok.PageCount != 3 || ok.ProviderCount != 1 || ok.ArtifactCount != 3 {
		t.Fatalf("verified evidence = %+v", ok)
	}
	failed := buildEvidence(m, []string{"missing"})
	if failed.Status != "failed" || problemError(failed.ProblemSummaries).Error() != "missing" {
		t.Fatalf("failed evidence = %+v", failed)
	}
}

func TestReadJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	var m manifest
	if err := readJSON([]byte(`{"schema_version":"v","id":"id"}`), &m); err != nil {
		t.Fatal(err)
	}
	if err := readJSON([]byte(`{"schema_version":"v","id":"id","extra":true}`), &m); err == nil {
		t.Fatal("expected unknown field to fail")
	}
	if err := readJSON([]byte(`{"schema_version":"v","id":"id"} {}`), &m); err == nil {
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
