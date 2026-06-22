package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunEvidenceScenariosReportsS3Publish(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.json")
	body := `{"steps":[{"id":"upload-dashboard-html","status":"passed"}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := runEvidenceScenarios(path)
	if len(got) != 1 || got[0].ID != "local.qa.s3_publish" || got[0].Status != "passed" {
		t.Fatalf("scenarios=%+v", got)
	}
}

func TestLoadLocalRunEvidenceReadsOpenRepairs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.json")
	body := `{"observed_at":"2026-06-22T01:00:00Z","expires_at":"2026-06-23T01:00:00Z","status":"passed","coverage_status":"partial","open_repairs":[{"provider_id":"cursor","class":"provider_auth_required","owner":"human","mode":"manual","summary":"login"}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := loadLocalRunEvidence(path)
	if !ok || got.ExpiresAt == "" || got.CoverageStatus != "partial" || got.OpenRepairs[0].ProviderID != "cursor" {
		t.Fatalf("run=%+v ok=%v", got, ok)
	}
}
