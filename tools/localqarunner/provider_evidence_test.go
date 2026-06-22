package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyProviderEvidenceMarksCoveragePartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "provider.json")
	body := `{"status":"partial","providers":[{"id":"cursor","repair":{"class":"provider_auth_required","owner":"human","mode":"manual","summary":"login required","suggested_command":"cursor-agent login"}}]}`
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
}
