package main

import "testing"

func TestBuildCoverageUsesExternalEvidence(t *testing.T) {
	rows, summary := buildCoverage(
		coverageManifest{Scenarios: []coverageScenario{{
			ID: "product.login", Evidence: "external",
		}}},
		providerEvidenceFile{},
		externalEvidenceFile{Scenarios: []externalScenario{{
			ID: "product.login", Status: "skipped",
			Evidence: "product.json", ExpiresAt: "2999-01-01T00:00:00Z",
			Repair: &repairEvidence{Class: "credential_required"},
		}}},
	)

	if summary.Skipped != 1 || rows[0].Repair.Class != "credential_required" {
		t.Fatalf("external coverage not applied: %+v %+v", summary, rows[0])
	}
	if rows[0].Evidence != "product.json" || rows[0].ExpiresAt == "" {
		t.Fatalf("external provenance missing: %+v", rows[0])
	}
}

func TestExternalScreenshotBecomesDashboardRelativeLink(t *testing.T) {
	rows, _ := buildCoverage(
		coverageManifest{Scenarios: []coverageScenario{{
			ID: "product.login", Evidence: "external",
		}}},
		providerEvidenceFile{},
		externalEvidenceFile{Scenarios: []externalScenario{{
			ID: "product.login", Status: "passed",
			Screenshot: ".riido-local/screenshots/ai-agent-product-acceptance/login.png",
		}}},
	)

	if rows[0].Screenshot != "screenshots/ai-agent-product-acceptance/login.png" {
		t.Fatalf("screenshot=%q", rows[0].Screenshot)
	}
}

func TestBuildCoverageUsesProviderProvenance(t *testing.T) {
	rows, _ := buildCoverage(
		coverageManifest{Scenarios: []coverageScenario{{
			ID: "provider.codex", Evidence: "provider", ProviderID: "codex",
		}}},
		providerEvidenceFile{
			ExpiresAt: "2999-01-01T00:00:00Z", EvidenceArtifact: "provider.json",
			Providers: []providerEvidence{{ID: "codex", IntegrationStatus: "passed"}},
		},
		externalEvidenceFile{},
	)
	if rows[0].Evidence != "provider.json" || rows[0].ExpiresAt == "" {
		t.Fatalf("provider provenance missing: %+v", rows[0])
	}
}
