package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPreviousCandidatesReadsValidEvidenceOnly(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.json")
	if got := loadPreviousCandidates(missing); got != nil {
		t.Fatalf("missing candidates=%+v", got)
	}
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := loadPreviousCandidates(bad); got != nil {
		t.Fatalf("invalid candidates=%+v", got)
	}
	good := filepath.Join(dir, "good.json")
	body := `{"closed_loop_candidates":[{"id":"candidate.auth"}]}`
	if err := os.WriteFile(good, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := loadPreviousCandidates(good)
	if len(got) != 1 || got[0].ID != "candidate.auth" {
		t.Fatalf("valid candidates=%+v", got)
	}
}

func TestProviderReadFailureRecordsAggregateError(t *testing.T) {
	step := providerReadFailure(os.ErrNotExist)
	if step.ID != "provider-evidence-aggregate" || step.Status != statusFailed {
		t.Fatalf("unexpected step=%+v", step)
	}
	if !strings.Contains(step.OutputTail, "file does not exist") {
		t.Fatalf("missing error tail=%q", step.OutputTail)
	}
}

func TestUploadDirMarksRecursiveUpload(t *testing.T) {
	got := uploadDir("screens", "source", "target/")
	if got.id != "upload-screens" || got.source != "source" || got.target != "target/" {
		t.Fatalf("unexpected upload=%+v", got)
	}
	if !got.recursive {
		t.Fatalf("upload dir should be recursive: %+v", got)
	}
}

func TestSyncFinalRunEvidenceFailsAtSyncBoundary(t *testing.T) {
	cfg := uploadTestConfig("", "", ".riido-local/coverage.json", "", "", "")
	evidencePath := writeUploadFixture(t, "run.json", "{}")
	cfg.runEvidence = &evidencePath
	original := runFinalSyncStep
	t.Cleanup(func() { runFinalSyncStep = original })
	runFinalSyncStep = func(root, id, exe string, args ...string) stepEvidence {
		return stepEvidence{ID: id, Status: statusFailed, OutputTail: "denied"}
	}
	err := syncFinalRunEvidence(".", cfg, "2026-06-22T00:00:00Z")
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("expected sync failure, got %v", err)
	}
}
