package main

import (
	"os"

	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

func providerObserved(provider provider, executable string) map[string]any {
	switch provider.ID {
	case "cursor":
		return cursorObserved(executable)
	case "openclaw":
		return openClawObserved()
	default:
		return nil
	}
}

func openClawObserved() map[string]any {
	return map[string]any{
		"integration_model_candidates": openclaw.IntegrationModelCandidates(),
		"model_override_env":           "RIIDO_OPENCLAW_INTEGRATION_MODEL",
	}
}

func cursorObserved(executable string) map[string]any {
	return map[string]any{
		"auth_preflight": cursorAuthPreflight{
			InteractiveLoginProbe: "cursor-agent about",
			InteractiveLoggedIn:   cursorInteractiveLoggedIn(executable),
			HeadlessAPIKeyEnv:     cursor.APIKeyEnv,
			HeadlessAPIKeyPresent: os.Getenv(cursor.APIKeyEnv) != "",
		},
	}
}
