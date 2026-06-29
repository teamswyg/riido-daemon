package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunEvidenceScenariosReportsS3Publish(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.json")
	body := `{"expires_at":"2999-01-01T00:00:00Z","steps":[{"id":"upload-dashboard-html","status":"passed"}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := runEvidenceScenarios(path)
	if len(got) != 1 || got[0].ID != "local.qa.s3_publish" || got[0].Status != "passed" {
		t.Fatalf("scenarios=%+v", got)
	}
	if got[0].Evidence != path || got[0].ExpiresAt == "" {
		t.Fatalf("run provenance missing: %+v", got[0])
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

func TestLoadLocalRunEvidenceReadsClosedLoopCandidates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.json")
	body := `{"observed_at":"2026-06-22T01:00:00Z","expires_at":"2026-06-23T01:00:00Z","status":"partial","coverage_status":"partial","closed_loop_candidates":[{"id":"coverage.product","source":"coverage","trigger":"coverage_not_passed","status":"promoted","summary":"missing","next_action":"promote","promoted":true,"age_hours":96,"stale_at":"2026-06-25T01:00:00Z","evidence_graph":{"decision":"promoted_to_closed_loop","next_loop":"closed-loop.coverage-product"}}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := loadLocalRunEvidence(path)
	if !ok || len(got.Candidates) != 1 || got.Candidates[0].ID != "coverage.product" {
		t.Fatalf("run=%+v ok=%v", got, ok)
	}
	if got.Candidates[0].AgeHours != 96 || got.Candidates[0].StaleAt == "" {
		t.Fatalf("candidate=%+v", got.Candidates[0])
	}
	if !got.Candidates[0].Promoted || got.Candidates[0].Graph.Decision != "promoted_to_closed_loop" ||
		got.Candidates[0].Graph.NextLoop != "closed-loop.coverage-product" {
		t.Fatalf("candidate=%+v", got.Candidates[0])
	}
}
