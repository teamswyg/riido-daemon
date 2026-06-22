package main

import "strings"

func classifyRepair(providerID, status, summary string, available bool) repair {
	if status == "passed" || status == "observed" {
		return repair{}
	}
	if !available {
		return repair{
			Class:   "provider_executable_missing",
			Owner:   "human",
			Mode:    "manual",
			Summary: "Install the provider CLI or set the provider override env.",
		}
	}
	text := strings.ToLower(summary)
	switch {
	case strings.Contains(text, "not logged in") || strings.Contains(text, "account missing"):
		return authRepair(providerID)
	case strings.Contains(text, "hard timeout") || strings.Contains(text, "timeout"):
		return providerTimeoutRepair()
	case strings.Contains(text, "config invalid") || strings.Contains(text, "invalid config"):
		return openClawConfigRepair()
	case localBackendUnavailable(text):
		return openClawBackendRepair()
	case strings.Contains(text, "completed without writing expected artifact"):
		return providerSideEffectRepair()
	default:
		return repair{
			Class:   "provider_integration_unclassified",
			Owner:   "engineer",
			Mode:    "manual",
			Summary: "Inspect failure_summary and add a more specific repair classifier.",
		}
	}
}

func providerTimeoutRepair() repair {
	return repair{
		Class:   "provider_timeout",
		Owner:   "engineer",
		Mode:    "manual",
		Summary: "Provider integration exceeded its hard timeout; inspect provider responsiveness or tune the acceptance timeout.",
	}
}

func providerSideEffectRepair() repair {
	return repair{
		Class:   "provider_side_effect_missing",
		Owner:   "engineer",
		Mode:    "manual",
		Summary: "Provider completed but did not perform the required filesystem side effect.",
	}
}

func localBackendUnavailable(text string) bool {
	return strings.Contains(text, "local model backend unavailable") ||
		strings.Contains(text, "connection refused by the provider endpoint") ||
		strings.Contains(text, "failovererror") ||
		strings.Contains(text, "provider ollama") ||
		strings.Contains(text, "all models failed") ||
		strings.Contains(text, "model-fallback") ||
		strings.Contains(text, "cooldown")
}

func authRepair(providerID string) repair {
	return repair{
		Class:   "provider_auth_required",
		Owner:   "human",
		Mode:    "manual",
		Summary: providerID + " requires an authenticated local provider session.",
	}
}
