package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReadEvidenceReportsReadAndDecodeErrors(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing.json")
	if _, err := readEvidence(missing); err == nil ||
		!strings.Contains(err.Error(), "read evidence") {
		t.Fatalf("expected read evidence error, got %v", err)
	}
	bad := filepath.Join(root, "bad.json")
	mustWrite(t, bad, "{")
	if _, err := readEvidence(bad); err == nil ||
		!strings.Contains(err.Error(), "decode evidence") {
		t.Fatalf("expected decode evidence error, got %v", err)
	}
}

func TestRunPropagatesReportWriteFailure(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeFixtureManifest(t, root)
	evidenceDir := filepath.Join(root, "out")
	mustMkdir(t, evidenceDir)
	mustWrite(t, filepath.Join(evidenceDir, "loop.json"), `{"status":"verified","problem_count":0}`)
	blocked := filepath.Join(root, "blocked")
	mustWrite(t, blocked, "file")
	err := run(options{
		Manifest:    manifestPath,
		EvidenceDir: evidenceDir,
		EvidenceOut: filepath.Join(blocked, "report.json"),
	})
	if err == nil || !strings.Contains(err.Error(), "write") {
		t.Fatalf("expected report write error, got %v", err)
	}
}

func TestBuildReportRecordsMalformedEvidenceAsFailedEvidence(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeFixtureManifest(t, root)
	m, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	evidenceDir := filepath.Join(root, "out")
	mustMkdir(t, evidenceDir)
	mustWrite(t, filepath.Join(evidenceDir, "loop.json"), "{")
	report := buildReport(evidenceDir, m)
	if report.Status != statusFailed || report.VerifiedCount != 0 {
		t.Fatalf("expected failed report for malformed evidence: %#v", report)
	}
	assertSelfProblem(t, report.Problems, "decode evidence")
}
