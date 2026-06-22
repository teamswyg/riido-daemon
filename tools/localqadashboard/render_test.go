package main

import (
	"strings"
	"testing"
)

func TestRenderDashboardIncludesFreshnessAndProviderStatus(t *testing.T) {
	html, err := renderDashboard(dashboardView{
		Evidence: providerEvidenceFile{
			Status:     "partial",
			ObservedAt: "2026-06-22T00:00:00Z",
			ExpiresAt:  "2026-06-23T00:00:00Z",
			Platform:   evidencePlatform{OS: "darwin", Arch: "arm64"},
			Providers:  []providerEvidence{{ID: "codex", Version: "codex-cli", IntegrationStatus: "passed"}},
		},
		Run: localRunEvidence{
			ObservedAt:     "2026-06-22T01:00:00Z",
			ExpiresAt:      "2026-06-23T01:00:00Z",
			Status:         "passed",
			CoverageStatus: "partial",
			OpenRepairs: []repairEvidence{{
				ProviderID: "cursor",
				Class:      "provider_auth_required",
				Owner:      "human",
				Mode:       "manual",
				Summary:    "login required",
			}},
		},
		CoverageRows: []coverageRow{{ID: "provider.codex", Title: "Codex", Status: "passed"}},
		CoverageSummary: coverageSummary{
			Total:  1,
			Passed: 1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Coverage Status", "provider_auth_required", "2026-06-23T01:00:00Z", "provider.codex", "passed"} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered dashboard missing %q", want)
		}
	}
	if strings.Contains(html, "2026-06-23T00:00:00Z") {
		t.Fatal("rendered dashboard used provider expiry instead of run expiry")
	}
}
