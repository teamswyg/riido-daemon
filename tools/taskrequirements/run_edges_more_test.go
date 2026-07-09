package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceForInvalidManifest(t *testing.T) {
	repo := t.TempDir()
	if err := writeJSON(filepath.Join(repo, "manifest.json"), Manifest{ID: "task-req"}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(repo, "evidence.json")
	err := run(options{Repo: repo, Manifest: "manifest.json", EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "missing schema_version") {
		t.Fatalf("expected validation error, got %v", err)
	}
	var ev Evidence
	raw, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatal(err)
	}
	if ev.ID != "task-req" || ev.ProblemCount == 0 {
		t.Fatalf("evidence=%+v", ev)
	}
}

func TestRunReturnsLoadAndEvidenceWriteErrors(t *testing.T) {
	repo := t.TempDir()
	if err := run(options{Repo: repo, Manifest: "missing.json"}); err == nil {
		t.Fatal("expected missing manifest error")
	}
	m := validTaskRequirementsManifest()
	if err := writeJSON(filepath.Join(repo, "manifest.json"), m); err != nil {
		t.Fatal(err)
	}
	err := run(options{Repo: repo, Manifest: "manifest.json", EvidenceOut: repo})
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("expected evidence write error, got %v", err)
	}
}

func TestNormalizeOptionsAndRepoPath(t *testing.T) {
	opts := normalizeOptions(options{})
	if opts.Repo != "." || opts.Manifest != defaultManifest {
		t.Fatalf("opts=%+v", opts)
	}
	abs := filepath.Join(t.TempDir(), "task.md")
	if got := repoPath("ignored", abs); got != abs {
		t.Fatalf("absolute path changed: %q", got)
	}
}

func validTaskRequirementsManifest() Manifest {
	return Manifest{
		SchemaVersion: "riido-task-requirements.v1",
		ID:            "task-req",
		Title:         "Task Requirements",
		GeneratedDoc:  "doc.md",
		Workflow:      "workflow.yml",
		Surfaces: []Surface{{
			Name: "mcp", CapabilityFlag: "SupportsMCP",
			SchedulingConstant: "SurfaceMCP", CandidateField: "SupportsMCP",
		}},
		Inputs:       []Input{{Name: "metadata.required_surfaces"}},
		SourceChecks: []SourceCheck{{Name: "src", File: "missing.go", Contains: "x"}},
		Assertions:   []string{"unknown surfaces fail closed"},
	}
}
