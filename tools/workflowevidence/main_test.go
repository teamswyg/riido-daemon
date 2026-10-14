package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkflowEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-evidence-out", out}); err != nil {
		t.Fatalf("run workflow evidence: %v", err)
	}
	var got evidence
	if err := readJSON(out, &got); err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if got.Status != "verified" || got.WorkflowCount == 0 || got.CoveredCount == 0 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if len(got.StatusCounts) == 0 {
		t.Fatalf("missing status counts: %+v", got)
	}
}

func readJSON(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
