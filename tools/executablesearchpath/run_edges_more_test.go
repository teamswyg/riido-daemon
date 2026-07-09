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
	manifestPath := filepath.Join(repo, "manifest.json")
	if err := writeJSON(manifestPath, Manifest{ID: "exec-path"}); err != nil {
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
	if ev.ID != "exec-path" || ev.ProblemCount == 0 {
		t.Fatalf("evidence=%+v", ev)
	}
}

func TestRunReturnsLoadAndEvidenceWriteErrors(t *testing.T) {
	repo := t.TempDir()
	if err := run(options{Repo: repo, Manifest: "missing.json"}); err == nil {
		t.Fatal("expected missing manifest error")
	}
	m := validExecutableSearchPathManifest()
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
	abs := filepath.Join(t.TempDir(), "file.md")
	if got := repoPath("ignored", abs); got != abs {
		t.Fatalf("absolute path changed: %q", got)
	}
}

func validExecutableSearchPathManifest() Manifest {
	return Manifest{
		SchemaVersion: "riido-executable-search-path.v1",
		ID:            "exec-path",
		Title:         "Executable Search Path",
		GeneratedDoc:  "doc.md",
		Workflow:      "workflow.yml",
		SearchOrder:   []SearchStep{{Name: "path"}},
		Rules:         []Rule{{Name: "rule"}},
		SourceChecks:  []SourceCheck{{Name: "src", File: "missing.go", Contains: "x"}},
	}
}
