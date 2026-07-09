package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextMapEvidenceAndProblemHelpers(t *testing.T) {
	m := manifest{
		ID: "context-map", GeneratedDoc: "docs/context.md",
		EvidenceArtifact: "evidence.json",
		Contexts:         []contextRow{{ID: "daemon"}},
		ACL:              aclFragment{Rows: []aclRow{{ACL: "desktop"}}},
		FigmaDaemon:      figmaFragment{Sections: []figmaSection{{Refs: []string{"figma"}}}},
		FigmaOnboarding:  onboardingFragment{Sections: []figmaSection{{Refs: []string{"figma"}}}},
	}
	ev := buildEvidence(m, map[string]string{"doc": "body"}, []string{"drift"})
	if ev.Status != "failed" || ev.ContextCount != 1 || ev.FigmaBoundaryCount != 2 {
		t.Fatalf("unexpected evidence=%+v", ev)
	}
	if statusFor(nil) != "verified" || !strings.Contains(problemError([]string{"a", "b"}).Error(), "a\nb") {
		t.Fatal("problem helpers should preserve status and messages")
	}
}

func TestContextMapIOWritesAndRejectsNonSemanticJSON(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "docs", "context.md")
	if err := writeText(textPath, "doc"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(textPath); err != nil || string(data) != "doc" {
		t.Fatalf("text=%q err=%v", data, err)
	}
	jsonPath := filepath.Join(dir, "evidence.json")
	if err := writeJSON(jsonPath, map[string]string{"id": "context"}); err != nil {
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
