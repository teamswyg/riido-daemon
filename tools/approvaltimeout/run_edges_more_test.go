package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceBeforeReturningProblems(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "internal/agentbridge/session/session_runner_timers.go", "package session\n")
	out := filepath.Join(repo, "evidence.json")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest, EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "source drift") {
		t.Fatalf("expected source drift, got %v", err)
	}
	var evidence Evidence
	if dataErr := readApprovalEvidence(out, &evidence); dataErr != nil {
		t.Fatal(dataErr)
	}
	if evidence.Status != "failed" || len(evidence.ProblemSummaries) == 0 {
		t.Fatalf("expected failed evidence, got %#v", evidence)
	}
}

func TestValidateReportsReferencedManifestLoadFailures(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	manifest := mustApprovalManifest(t, repo)
	manifest.Sources.SemanticActivityManifest = "missing.json"
	problems, checks, sources := validate(repo, manifest)
	if len(checks) != 0 || len(sources) != 0 {
		t.Fatalf("checks=%#v sources=%#v", checks, sources)
	}
	assertApprovalProblem(t, problems, "missing.json")
	manifest.Sources.SemanticActivityManifest = "../bad.json"
	problems, _, _ = validate(repo, manifest)
	assertApprovalProblem(t, problems, "unsafe repo path")
}

func TestIOReportsUnsafePathsAndWriteFailures(t *testing.T) {
	repo := t.TempDir()
	if _, err := loadJSON[Manifest](repo, "../bad.json"); err == nil {
		t.Fatal("expected unsafe load path")
	}
	if _, err := readSource(repo, "../bad.go"); err == nil {
		t.Fatal("expected unsafe source path")
	}
	if err := writeJSON(filepath.Join(repo, "missing", "out.json"), map[string]string{"ok": "yes"}); err == nil {
		t.Fatal("expected missing directory write failure")
	}
}

func readApprovalEvidence(path string, out *Evidence) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
