package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadWorkflowSourcesRejectsInvalidFragments(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	sourcePath := filepath.Join(dir, "sources", "empty.json")
	mustWrite(t, manifestPath, `{"schema_version":"riido-daemon-package-workflow-evidence.v1","id":"m","loop_source":"loop","workflow_sources":["sources/empty.json"],"workflows":[]}`)
	mustWrite(t, sourcePath, `{"schema_version":"riido-daemon-package-workflow-evidence.v1","id":"fragment","workflows":[]}`)

	_, err := loadManifest(manifestPath)
	if err == nil || !strings.Contains(err.Error(), "invalid workflow source") {
		t.Fatalf("expected invalid source error, got %v", err)
	}
}

func TestLoadManifestReportsMissingSourceFile(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	mustWrite(t, manifestPath, `{"schema_version":"riido-daemon-package-workflow-evidence.v1","id":"m","loop_source":"loop","workflow_sources":["missing.json"],"workflows":[]}`)

	_, err := loadManifest(manifestPath)
	if err == nil || !strings.Contains(err.Error(), "read manifest") {
		t.Fatalf("expected missing source read error, got %v", err)
	}
}

func TestWriteJSONReportsEvidenceDirAndWriteErrors(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "not-dir"), "file")
	err := writeJSON(filepath.Join(dir, "not-dir", "evidence.json"), evidence{})
	if err == nil || !strings.Contains(err.Error(), "create evidence dir:") {
		t.Fatalf("expected evidence dir error, got %v", err)
	}

	err = writeJSON(dir, evidence{})
	if err == nil || !strings.Contains(err.Error(), "write evidence:") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestRunRequiresWorkflowAndEvidenceOut(t *testing.T) {
	if err := run(nil); err == nil || !strings.Contains(err.Error(), "-workflow and -evidence-out are required") {
		t.Fatalf("expected missing flag error, got %v", err)
	}
}
