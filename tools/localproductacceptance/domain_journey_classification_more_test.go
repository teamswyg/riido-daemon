package main

import "testing"

func TestDomainEnvironmentAndVerificationClassification(t *testing.T) {
	for _, tc := range []struct {
		name      string
		riidoHost string
		agentHost string
		want      string
	}{
		{"staging", "https://staging.api.riido.io", "", "staging"},
		{"production", "https://api.riido.io", "https://production.ai-api.riido.io", "production"},
		{"development", "", "https://development.ai-api.riido.io", "development"},
		{"custom", "https://example.invalid", "", "custom"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := domainRemoteEnvironment(tc.riidoHost, tc.agentHost); got != tc.want {
				t.Fatalf("environment=%q want %q", got, tc.want)
			}
		})
	}
	for _, base := range []string{"http://localhost:3000", "http://127.0.0.1:5173"} {
		if got := domainVerificationSource(base); got != "local" {
			t.Fatalf("verification source=%q for %s", got, base)
		}
	}
	if got := domainVerificationSource("https://app.riido.io"); got != "custom" {
		t.Fatalf("verification source=%q", got)
	}
}

func TestDomainJourneySummarySkipsOutsideStagingLocalPair(t *testing.T) {
	cfg := domainJourneyTestConfig("missing.json")
	riidoHost, agentHost, base := "https://api.riido.io", "https://ai-api.riido.io", "https://app.riido.io"
	cfg.riidoAPIHost = &riidoHost
	cfg.agentHost = &agentHost
	cfg.baseURL = &base

	row := domainJourneySummary(cfg, domainFixtureCache{}, nil, domainEntityDefs())
	if row.Status != statusSkipped || row.Repair == nil {
		t.Fatalf("row=%+v", row)
	}
	if row.Observed["remote_environment"] != "custom" ||
		row.Observed["verification_source"] != "custom" {
		t.Fatalf("observed=%+v", row.Observed)
	}
}
