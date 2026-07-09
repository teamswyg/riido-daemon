package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceBeforeReturningSourceProblems(t *testing.T) {
	dir := t.TempDir()
	policyPath, err := filepath.Abs("../../internal/hostintegration/privacy_metadata_allowlist.riido.json")
	if err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{
		SchemaVersion: "v", ID: "privacy", Title: "Privacy",
		GeneratedDoc: "privacy.md", Workflow: "ci", PolicyArtifact: policyPath,
		SourceChecks: []SourceCheck{{Name: "missing", File: "source.go", Contains: "needle"}},
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	if err := writeJSON(manifestPath, manifest); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "evidence.json")
	err = run(options{Repo: dir, Manifest: "manifest.json", EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "missing:") {
		t.Fatalf("expected source problem, got %v", err)
	}
	var evidence Evidence
	data, readErr := os.ReadFile(out)
	if readErr != nil || json.Unmarshal(data, &evidence) != nil {
		t.Fatalf("failed evidence missing: err=%v body=%s", readErr, data)
	}
	if len(evidence.Problems) == 0 || len(evidence.SourceChecks) != 1 ||
		evidence.SourceChecks[0].OK {
		t.Fatalf("expected failed evidence, got %#v", evidence)
	}
	if len(evidence.ShapeChecks) != 3 {
		t.Fatalf("shape evidence missing: %#v", evidence.ShapeChecks)
	}
}

func TestCheckDocReportsMissingAndDrift(t *testing.T) {
	dir := t.TempDir()
	opts := options{Repo: dir, CheckDoc: true}
	problems := checkDoc(opts, "privacy.md", "expected")
	if len(problems) == 0 || !strings.Contains(problems[0].Message, "no such file") {
		t.Fatalf("expected missing doc problem, got %#v", problems)
	}
	if err := writeText(filepath.Join(dir, "privacy.md"), "old"); err != nil {
		t.Fatal(err)
	}
	problems = checkDoc(opts, "privacy.md", "expected")
	if len(problems) == 0 || !strings.Contains(problems[0].Message, "generated doc drift") {
		t.Fatalf("expected drift problem, got %#v", problems)
	}
}
