package main

import "github.com/teamswyg/riido-daemon/internal/provider/openclaw"

func providerObserved(provider provider) map[string]any {
	if provider.ID != "openclaw" {
		return nil
	}
	return map[string]any{
		"integration_model_candidates": openclaw.IntegrationModelCandidates(),
		"model_override_env":           "RIIDO_OPENCLAW_INTEGRATION_MODEL",
	}
}
