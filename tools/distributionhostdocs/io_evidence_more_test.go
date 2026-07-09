package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDistributionHostEvidenceAndProblemHelpers(t *testing.T) {
	m := manifest{
		ID: "distribution-host", GeneratedDoc: "docs/root.md",
		EvidenceArtifact: "evidence.json",
		Pages:            []page{{GeneratedDoc: "docs/page.md"}},
		Assertions:       []string{"must hold"},
	}
	checks := []sourceCheckResult{{Name: "anchor", File: "README.md", Passed: true}}
	ev := buildEvidence(m, checks, []string{"drift"})
	if ev.Status != "failed" || ev.PageCount != 2 || ev.GeneratedDocs[1] != "docs/page.md" {
		t.Fatalf("unexpected evidence=%+v", ev)
	}
	if statusFor(nil) != "verified" || !strings.Contains(problemError([]string{"a", "b"}).Error(), "a\nb") {
		t.Fatal("problem helpers should preserve status and messages")
	}
}

func TestDistributionHostIOWritesAndRejectsNonSemanticJSON(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "docs", "host.md")
	if err := writeText(textPath, "doc"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(textPath); err != nil || string(data) != "doc" {
		t.Fatalf("text=%q err=%v", data, err)
	}
	jsonPath := filepath.Join(dir, "evidence.json")
	if err := writeJSON(jsonPath, map[string]string{"id": "host"}); err != nil {
		t.Fatal(err)
	}
	var m manifest
	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"unknown":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
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
