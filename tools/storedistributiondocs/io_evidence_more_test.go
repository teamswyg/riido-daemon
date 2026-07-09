package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.json")
	if err := os.WriteFile(valid, []byte(`{"schema_version":"v","id":"id"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := readJSON(valid, &m); err != nil {
		t.Fatal(err)
	}
	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"schema_version":"v","id":"id","extra":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(unknown, &m); err == nil {
		t.Fatal("expected unknown field to fail")
	}
	trailing := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{"schema_version":"v","id":"id"} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(trailing, &m); err == nil {
		t.Fatal("expected trailing JSON value to fail")
	}
}

func TestEvidenceValidationAndProblemStatus(t *testing.T) {
	docs := map[string]string{"b.md": "B", "a.md": "A"}
	c := contract{Channels: []channel{{ID: "prod"}, {ID: "beta"}}}
	ev := buildEvidence(manifest{ID: "store", EvidenceArtifact: "evidence"}, c, docs, nil)
	if ev.Status != "verified" || strings.Join(ev.GeneratedDocs, ",") != "a.md,b.md" {
		t.Fatalf("verified evidence = %+v", ev)
	}
	failed := buildEvidence(manifest{ID: "store"}, c, docs, []string{"bad"})
	if failed.Status != "failed" || problemError(failed.ProblemSummaries).Error() != "bad" {
		t.Fatalf("failed evidence = %+v", failed)
	}
	problems := validateInputs(manifest{}, contract{})
	if len(problems) != 5 {
		t.Fatalf("problems = %+v", problems)
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
