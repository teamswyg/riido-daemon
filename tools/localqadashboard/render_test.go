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
			ExpiresAt:  "2999-06-23T00:00:00Z",
			Platform:   evidencePlatform{OS: "darwin", Arch: "arm64"},
			Providers:  []providerEvidence{{ID: "codex", Version: "codex-cli", IntegrationStatus: "passed"}},
		},
		Run: localRunEvidence{
			ObservedAt:     "2026-06-22T01:00:00Z",
			ExpiresAt:      "2999-06-23T01:00:00Z",
			Status:         "passed",
			CoverageStatus: "partial",
			OpenRepairs: []repairEvidence{{
				ProviderID:       "cursor",
				Class:            "provider_auth_required",
				Owner:            "human",
				Mode:             "manual",
				Summary:          "login required",
				SuggestedCommand: "cursor-agent login",
			}},
			Candidates: []closedLoopCandidate{{
				ID:         "coverage.product",
				Source:     "coverage",
				Trigger:    "coverage_not_passed",
				Status:     "candidate",
				Summary:    "missing product outcome",
				NextAction: "promote to verifier",
				AgeHours:   12,
				StaleAt:    "2026-06-25T00:00:00Z",
			}},
		},
		CoverageRows: []coverageRow{{
			ID:         "figma.onboarding",
			Title:      "Onboarding",
			Status:     "passed",
			Evidence:   "figma.json",
			ExpiresAt:  "2999-06-23T01:00:00Z",
			Screenshot: "screenshots/ai-agent-product-acceptance/figma-onboarding.png",
		}},
		CoverageSummary: coverageSummary{
			Total:  1,
			Passed: 1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Coverage Status",
		"freshness-status",
		`data-expires="2999-06-23T01:00:00Z"`,
		">fresh</div>",
		"Deployment Gate",
		"blocked",
		"-strict-coverage",
		"provider_auth_required",
		"cursor-agent login",
		"Closed-Loop Candidates",
		"coverage.product",
		"12h",
		"stale 2026-06-25T00:00:00Z",
		"promote to verifier",
		"figma.onboarding",
		"figma.json",
		"expires 2999-06-23T01:00:00Z",
		`<img class="shot"`,
		"passed",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered dashboard missing %q", want)
		}
	}
	if strings.Contains(html, "2999-06-23T00:00:00Z") {
		t.Fatal("rendered dashboard used provider expiry instead of run expiry")
	}
}
