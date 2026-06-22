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
		CoverageRows: []coverageRow{{ID: "provider.codex", Title: "Codex", Status: "passed"}},
		CoverageSummary: coverageSummary{
			Total:  1,
			Passed: 1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"partial", "2026-06-23T00:00:00Z", "provider.codex", "passed"} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered dashboard missing %q", want)
		}
	}
}
