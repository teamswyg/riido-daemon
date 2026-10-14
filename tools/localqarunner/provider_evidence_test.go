package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyProviderEvidenceMarksCoveragePartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "provider.json")
	body := `{"status":"partial","providers":[{"id":"cursor","available":true,"version":"2026.06","integration_status":"skipped","observed":{"auth_preflight":{"headless_api_key_present":false}},"repair":{"class":"provider_auth_required","owner":"human","mode":"manual","summary":"login required","suggested_command":"cursor-agent login"}}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config{providerEvidence: &path}
	evidence := runEvidence{Status: statusPassed, CoverageStatus: statusPassed}
	if err := applyProviderEvidence("/repo", cfg, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.CoverageStatus != statusPartial || evidence.ProviderStatus != statusPartial {
		t.Fatalf("coverage=%q provider=%q", evidence.CoverageStatus, evidence.ProviderStatus)
	}
	if len(evidence.OpenRepairs) != 1 || evidence.OpenRepairs[0].SuggestedCommand == "" {
		t.Fatalf("repairs=%+v", evidence.OpenRepairs)
	}
	if evidence.OpenRepairs[0].ProviderID != "cursor" {
		t.Fatalf("repair provider id missing: %+v", evidence.OpenRepairs)
	}
	if len(evidence.ProviderSummary) != 1 {
		t.Fatalf("provider summary=%+v", evidence.ProviderSummary)
	}
	summary := evidence.ProviderSummary[0]
	if !summary.Available || summary.Version != "2026.06" || summary.IntegrationStatus != "skipped" {
		t.Fatalf("provider summary=%+v", summary)
	}
	if summary.Repair == nil || summary.Repair.ProviderID != "cursor" {
		t.Fatalf("provider repair=%+v", summary.Repair)
	}
	if len(summary.Observed) == 0 {
		t.Fatal("observed provider evidence missing")
	}
	if len(evidence.ClosedLoops) != 0 {
		t.Fatalf("manual repair became closed-loop candidate: %+v", evidence.ClosedLoops)
	}
}
