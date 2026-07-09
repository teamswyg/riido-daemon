package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceForInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	mustRoadmapWrite(t, filepath.Join(dir, ".github", "workflows", "qa.yml"), "workflow")
	manifestPath := "roadmap.json"
	m := manifest{
		SchemaVersion:    manifestSchema,
		ID:               "roadmap",
		Title:            "Roadmap",
		GeneratedDoc:     "roadmap.md",
		Workflow:         ".github/workflows/qa.yml",
		EvidenceArtifact: "roadmap-docs",
		Questions: []question{
			{ID: "bad", Area: "Scope", Question: "?", CurrentHandling: "manual"},
		},
		Assertions: []string{"registry owns generated docs"},
	}
	if err := writeJSON(repoPath(dir, manifestPath), m); err != nil {
		t.Fatal(err)
	}
	evidencePath := filepath.Join(dir, "out", "evidence.json")
	err := run(options{Repo: dir, Manifest: manifestPath, EvidenceOut: evidencePath})
	if err == nil || !strings.Contains(err.Error(), "question id must start with Q-") {
		t.Fatalf("expected invalid manifest error, got %v", err)
	}
	var ev evidence
	if err := readJSON(evidencePath, &ev); err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if ev.Status != "failed" || ev.QuestionCount != 1 || len(ev.ProblemSummaries) == 0 {
		t.Fatalf("unexpected evidence=%+v", ev)
	}
}

func TestRunReturnsLoadAndEvidenceWriteErrors(t *testing.T) {
	dir := t.TempDir()
	if err := run(options{Repo: dir, Manifest: "missing.json"}); err == nil {
		t.Fatal("expected missing manifest error")
	}
	mustRoadmapWrite(t, filepath.Join(dir, ".github", "workflows", "qa.yml"), "workflow")
	m := manifest{
		SchemaVersion:    manifestSchema,
		ID:               "roadmap",
		Title:            "Roadmap",
		GeneratedDoc:     "roadmap.md",
		Workflow:         ".github/workflows/qa.yml",
		EvidenceArtifact: "roadmap-docs",
		Questions: []question{
			{ID: "Q-1", Area: "Scope", Question: "?", CurrentHandling: "manual"},
		},
		Assertions: []string{"registry owns generated docs"},
	}
	if err := writeJSON(repoPath(dir, "roadmap.json"), m); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(dir, "blocker")
	mustRoadmapWrite(t, blocker, "file")
	err := run(options{
		Repo:        dir,
		Manifest:    "roadmap.json",
		EvidenceOut: filepath.Join(blocker, "evidence.json"),
	})
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("expected evidence write error, got %v", err)
	}
}
