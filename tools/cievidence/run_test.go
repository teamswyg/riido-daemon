package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCIEvidenceVerifiesWorkflowCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ci.yml")
	writeWorkflow(t, path, "go list -m all\ngo test ./...\n")
	report := buildEvidence(options{Workflow: path, ID: "ci"}, mustRead(t, path))
	if report.Status != "verified" || len(report.Required) != 2 {
		t.Fatalf("report = %+v", report)
	}
}

func TestCIEvidenceRejectsMissingCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "go-ci.yml")
	writeWorkflow(t, path, "go test ./...\n")
	report := buildEvidence(options{Workflow: path, ID: "go-ci"}, mustRead(t, path))
	if report.Status != "failed" || len(report.Problems) == 0 {
		t.Fatalf("report = %+v", report)
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
