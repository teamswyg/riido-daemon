package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunBuildsDashboardAndCoverageSnapshot(t *testing.T) {
	dir := t.TempDir()
	providerPath := filepath.Join(dir, "provider.json")
	manifestPath := filepath.Join(dir, "coverage.json")
	outputPath := filepath.Join(dir, "out", "dashboard.html")
	snapshotPath := filepath.Join(dir, "out", "coverage-snapshot.json")
	mustWriteDashboardJSON(t, providerPath, providerEvidenceFile{
		Status: "passed", ObservedAt: "2026-01-01T00:00:00Z",
		ExpiresAt: "2999-01-01T00:00:00Z",
		Platform:  evidencePlatform{OS: "darwin", Arch: "arm64"},
		Providers: []providerEvidence{{ID: "codex", IntegrationStatus: "passed"}},
	})
	mustWriteDashboardJSON(t, manifestPath, coverageManifest{
		ID: "coverage",
		Scenarios: []coverageScenario{{
			ID: "provider.codex", Title: "Codex", Tier: "daily",
			Surface: "provider", Evidence: "provider", ProviderID: "codex",
		}},
	})
	if err := run(providerPath, "", "", "", "", "", manifestPath, outputPath, snapshotPath); err != nil {
		t.Fatalf("run dashboard: %v", err)
	}
	html, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read dashboard: %v", err)
	}
	if len(html) == 0 {
		t.Fatal("dashboard HTML should not be empty")
	}
	var snapshot coverageSnapshot
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snapshot.Summary.Total != 1 || snapshot.Rows[0].Status != "passed" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestReleaseEvidenceScenariosOptionalAndInvalid(t *testing.T) {
	if got := releaseEvidenceScenarios(""); got != nil {
		t.Fatalf("empty release evidence should be nil: %#v", got)
	}
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := releaseEvidenceScenarios(path); got != nil {
		t.Fatalf("invalid release evidence should be nil: %#v", got)
	}
}

func mustWriteDashboardJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
