package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceForMissingArtifact(t *testing.T) {
	root := t.TempDir()
	if err := writeText(repoPath(root, "workflow.yml"), "name: provider migration\n"); err != nil {
		t.Fatal(err)
	}
	m := validProviderMigrationManifest()
	manifestPath := "manifest.json"
	if err := writeJSON(repoPath(root, manifestPath), m); err != nil {
		t.Fatal(err)
	}
	evidencePath := filepath.Join(root, "out", "evidence.json")
	err := run(options{Repo: root, Manifest: manifestPath, EvidenceOut: evidencePath})
	if err == nil || !strings.Contains(err.Error(), "missing artifact \"missing.md\"") {
		t.Fatalf("expected missing artifact error, got %v", err)
	}
	var ev evidence
	data, readErr := os.ReadFile(evidencePath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err := readJSON(data, &ev); err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if ev.Status != "failed" || ev.PageCount != 2 || len(ev.ProblemSummaries) == 0 {
		t.Fatalf("unexpected evidence=%+v", ev)
	}
}

func TestRunReturnsLoadAndEvidenceWriteErrors(t *testing.T) {
	root := t.TempDir()
	if err := run(options{Repo: root, Manifest: "missing.json"}); err == nil {
		t.Fatal("expected missing manifest error")
	}
	if err := writeText(repoPath(root, "workflow.yml"), "name: provider migration\n"); err != nil {
		t.Fatal(err)
	}
	m := validProviderMigrationManifest()
	m.Pages[0].Artifacts = nil
	if err := writeJSON(repoPath(root, "manifest.json"), m); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(root, "blocker")
	if err := writeText(blocker, "file"); err != nil {
		t.Fatal(err)
	}
	err := run(options{
		Repo:        root,
		Manifest:    "manifest.json",
		EvidenceOut: filepath.Join(blocker, "evidence.json"),
	})
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("expected evidence write error, got %v", err)
	}
	if got := mustExist(root, ""); len(got) != 0 {
		t.Fatalf("empty path should pass: %v", got)
	}
	if got := mustExist(root, "missing.md"); len(got) != 1 {
		t.Fatalf("missing path should fail: %v", got)
	}
}
