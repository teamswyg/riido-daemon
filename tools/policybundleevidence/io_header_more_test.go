package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritersReportParentAndMarshalErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "not-dir"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := writeJSON(filepath.Join(root, "not-dir", "evidence.json"), map[string]string{"ok": "true"})
	if err == nil || !strings.Contains(err.Error(), "prepare evidence dir:") {
		t.Fatalf("expected evidence dir problem, got %v", err)
	}
	err = writeJSON(filepath.Join(root, "evidence.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "encode evidence:") {
		t.Fatalf("expected evidence encoding problem, got %v", err)
	}
	err = writeText(filepath.Join(root, "not-dir", "doc.md"), "doc")
	if err == nil || !strings.Contains(err.Error(), "prepare text dir:") {
		t.Fatalf("expected text dir problem, got %v", err)
	}
}

func TestRepoPathKeepsAbsoluteAndRootsRelativePaths(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "manifest.json")
	if got := repoPath("/repo", abs); got != abs {
		t.Fatalf("absolute path changed: %q", got)
	}
	if got := repoPath("/repo", "docs/policy.md"); got != filepath.Join("/repo", "docs", "policy.md") {
		t.Fatalf("relative path not rooted: %q", got)
	}
}

func TestValidateHeaderCoversValidAndMissingFields(t *testing.T) {
	valid := Manifest{
		SchemaVersion:    "riido-policy-bundle-loader-evidence.v1",
		ID:               "policy",
		Title:            "Policy",
		GeneratedDoc:     "doc.md",
		Workflow:         "workflow.yml",
		EvidenceArtifact: "artifact",
		Purpose:          "purpose",
		Facts:            []Fact{{Name: "f"}},
		Boundaries:       []Boundary{{Name: "b"}},
		SourceChecks:     []SourceCheck{{Name: "s"}},
	}
	if problems := validateHeader(valid); len(problems) != 0 {
		t.Fatalf("valid header should pass: %v", problems)
	}

	invalid := valid
	invalid.SchemaVersion = "old"
	invalid.ID = ""
	invalid.Facts = nil
	problems := strings.Join(messages(validateHeader(invalid)), "\n")
	for _, wanted := range []string{
		"schema_version must be",
		"id, title, generated_doc",
		"facts, boundaries, and source_checks",
	} {
		if !strings.Contains(problems, wanted) {
			t.Fatalf("missing %q in %q", wanted, problems)
		}
	}
}
