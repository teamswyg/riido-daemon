package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCIEvidenceVerifiesWorkflowCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ci.yml")
	writeWorkflow(t, path, "go list -m all\ngo test ./...\n")
	report := buildEvidence(testManifest(path), testSpec(path), mustRead(t, path))
	if report.Status != "verified" || len(report.Required) != 2 {
		t.Fatalf("report = %+v", report)
	}
	if report.ProblemCount != 0 || report.Problems == nil {
		t.Fatalf("problem evidence = %+v", report)
	}
}

func TestCIEvidenceRejectsMissingCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ci.yml")
	writeWorkflow(t, path, "go test ./...\n")
	report := buildEvidence(testManifest(path), testSpec(path), mustRead(t, path))
	if report.Status != "failed" || report.ProblemCount == 0 {
		t.Fatalf("report = %+v", report)
	}
}

func TestFindWorkflowRejectsIDMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ci.yml")
	_, err := findWorkflow(testManifest(path), path, "other")
	if err == nil {
		t.Fatal("expected id mismatch")
	}
}

func testManifest(path string) manifest {
	return manifest{
		SchemaVersion: manifestSchema,
		ID:            "daemon-ci-evidence",
		LoopSource:    "loops/ci.json",
		Workflows:     []workflowSpec{testSpec(path)},
	}
}

func testSpec(path string) workflowSpec {
	return workflowSpec{
		ID:               "ci",
		Workflow:         path,
		EvidenceArtifact: "ci-evidence",
		RequiredCommands: []string{"go list -m all", "go test ./..."},
	}
}

func writeWorkflow(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
