package main

import "testing"

func TestCoverageSourceFallbacksRespectExistingValues(t *testing.T) {
	provider := withProviderSource(providerEvidenceFile{
		ExpiresAt: "2999-01-01T00:00:00Z",
		Providers: []providerEvidence{{ID: "codex"}},
	}, "provider.json")
	if provider.Providers[0].EvidenceArtifact != "provider.json" ||
		provider.Providers[0].ExpiresAt == "" {
		t.Fatalf("provider source missing: %#v", provider)
	}
	external := withExternalSource(externalEvidenceFile{
		ExpiresAt: "2999-01-01T00:00:00Z",
		Scenarios: []externalScenario{
			{ID: "a"},
			{ID: "b", Evidence: "custom.json", ExpiresAt: "2998-01-01T00:00:00Z"},
		},
	}, "external.json")
	if external.Scenarios[0].Evidence != "external.json" || external.Scenarios[0].ExpiresAt == "" {
		t.Fatalf("default source missing: %#v", external)
	}
	if external.Scenarios[1].Evidence != "custom.json" ||
		external.Scenarios[1].ExpiresAt != "2998-01-01T00:00:00Z" {
		t.Fatalf("existing source should be preserved: %#v", external)
	}
	rows := withFallbackExpiry([]coverageRow{
		{ID: "a", Evidence: "external.json"},
		{ID: "b", Evidence: "external.json", ExpiresAt: "custom"},
		{ID: "c"},
	}, "fallback")
	if rows[0].ExpiresAt != "fallback" || rows[1].ExpiresAt != "custom" || rows[2].ExpiresAt != "" {
		t.Fatalf("unexpected fallback expiry: %#v", rows)
	}
}
