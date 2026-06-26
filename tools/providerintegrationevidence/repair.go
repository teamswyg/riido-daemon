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
	case providerToolApprovalMissing(providerID, text):
		return providerToolApprovalMissingRepair()
	case strings.Contains(text, "hard timeout") || strings.Contains(text, "timeout"):
		return providerTimeoutRepair()
	case strings.Contains(text, "config invalid") || strings.Contains(text, "invalid config"):
		return openClawConfigRepair()
	case localBackendUnavailable(text):
		return openClawBackendRepair()
	case strings.Contains(text, "completed without writing expected artifact"):
		if providerID == "openclaw" {
			return openClawModelConfigRepair()
		}
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

func providerToolApprovalMissing(providerID, text string) bool {
	if providerID != "claude" {
		return false
	}
	return strings.Contains(text, "execution permission") ||
		strings.Contains(text, "command execution permission") ||
		strings.Contains(text, "approval count 0") ||
		strings.Contains(text, "tool approval") ||
		strings.Contains(text, "permission limited") ||
		strings.Contains(text, "실행 권한") ||
		strings.Contains(text, "승인")
}

func providerToolApprovalMissingRepair() repair {
	return repair{
		Class:            "provider_tool_approval_missing",
		Owner:            "engineer",
		Mode:             "candidate_auto",
		Summary:          "Claude command permission evidence or conversational approval text did not produce a Riido tool approval request.",
		SuggestedCommand: "capture redacted Claude permission/conversation/control_request evidence, verify /tool-approvals is non-empty, and bind chat approval to the tool approval decision flow",
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

func authRepair(providerID string) repair {
	rep := repair{
		Class:   "provider_auth_required",
		Owner:   "human",
		Mode:    "manual",
		Summary: providerID + " requires an authenticated local provider session.",
	}
	if providerID == "cursor" {
		rep.Summary = "cursor requires cursor-agent login or CURSOR_API_KEY."
		rep.SuggestedCommand = "cursor-agent login && " +
			"AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/cursor -race -count=1 -run TestIntegration -v; " +
			"# headless: CURSOR_API_KEY=<key> AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/cursor -race -count=1 -run TestIntegration -v"
	}
	return rep
}
