package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyCoverageEvidenceEmbedsRowsAndStatus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.json")
	body := `{"summary":{"total":2,"passed":1,"skipped":1},"rows":[` +
		`{"id":"provider.cursor","title":"Cursor","status":"skipped",` +
		`"evidence":"provider.json","expires_at":"2999-01-01T00:00:00Z"}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config{coverageEvidence: &path}
	evidence := runEvidence{CoverageStatus: statusPassed}
	if err := applyCoverageEvidence(".", cfg, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.CoverageStatus != statusPartial || evidence.Coverage == nil {
		t.Fatalf("coverage=%+v", evidence)
	}
	if len(evidence.Coverage.Rows) != 1 || evidence.Coverage.Rows[0].ID != "provider.cursor" {
		t.Fatalf("rows=%+v", evidence.Coverage.Rows)
	}
	if evidence.Coverage.Rows[0].Evidence == "" || evidence.Coverage.Rows[0].ExpiresAt == "" {
		t.Fatalf("row provenance missing: %+v", evidence.Coverage.Rows[0])
	}
}

func TestCoverageStatusFromSummary(t *testing.T) {
	if coverageStatus(runCoverageSummary{Total: 1, Passed: 1}) != statusPassed {
		t.Fatal("expected passed")
	}
	if coverageStatus(runCoverageSummary{Total: 1, Failed: 1}) != statusFailed {
		t.Fatal("expected failed")
	}
	if coverageStatus(runCoverageSummary{Total: 2, Passed: 1}) != statusPartial {
		t.Fatal("expected partial")
	}
}
