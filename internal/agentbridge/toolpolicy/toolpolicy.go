// Package toolpolicy owns the C4 adapter/runtime mapping from provider-neutral
// tool references to C7 policy decisions. It does not parse provider raw
// schemas, execute provider processes, or own the C7 policy matrix itself.
package toolpolicy

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

// PolicyAutoApprover returns a session AutoApprover backed by the active C7
// policy bundle. Unknown or unclassified tools remain on the human approval
// path; only an explicit policy allow can auto-approve.
func PolicyAutoApprover(bundle policy.PolicyBundle, tier policy.TrustTier) agentbridge.AutoApprover {
	return func(tool agentbridge.ToolRef) bool {
		decision, ok := DecisionForTool(bundle, tier, tool)
		return ok && decision.Action == policy.ToolUseActionAllow
	}
}

// PolicyToolStartGate returns a session ToolStartGate backed by the active C7
// policy bundle. Only classified risk surfaces can block; unclassified tools
// keep running because the current classifier cannot prove they are risky.
func PolicyToolStartGate(bundle policy.PolicyBundle, tier policy.TrustTier) agentbridge.ToolStartGate {
	return func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
		decision, ok := DecisionForStartedTool(bundle, tier, tool)
		if !ok || decision.Action == policy.ToolUseActionAllow {
			return agentbridge.ToolStartDecision{}
		}
		return agentbridge.ToolStartDecision{
			Block:  true,
			Code:   decision.Code,
			Reason: decision.Reason,
		}
	}
}

// DecisionForTool classifies tool and evaluates the C7 ToolUseSecurityGate
// decision for provider approval flows.
func DecisionForTool(bundle policy.PolicyBundle, tier policy.TrustTier, tool agentbridge.ToolRef) (policy.ToolUseDecision, bool) {
	surface, ok := ClassifyToolUseSurface(tool)
	if !ok {
		return policy.ToolUseDecision{}, false
	}
	return policy.EvaluateToolUseWithBundle(bundle, policy.ToolUseInput{
		TrustTier:              tier,
		Surface:                surface,
		HumanApprovalAvailable: true,
	}), true
}

// DecisionForStartedTool classifies tool and evaluates the C7
// ToolUseSecurityGate decision for ToolCallStarted events where the provider
// has not offered a human approval round-trip.
func DecisionForStartedTool(bundle policy.PolicyBundle, tier policy.TrustTier, tool agentbridge.ToolRef) (policy.ToolUseDecision, bool) {
	surface, ok := ClassifyToolUseSurface(tool)
	if !ok {
		return policy.ToolUseDecision{}, false
	}
	return policy.EvaluateToolUseWithBundle(bundle, policy.ToolUseInput{
		TrustTier:              tier,
		Surface:                surface,
		HumanApprovalAvailable: false,
	}), true
}

// ClassifyToolUseSurface maps provider-neutral tool labels and redacted args
// into known C7 tool-use risk surfaces. It is intentionally conservative:
// absence of enough signal means "do not auto-approve".
func ClassifyToolUseSurface(tool agentbridge.ToolRef) (policy.ToolUseSurface, bool) {
	kind := normalizeToolToken(tool.Kind)
	name := normalizeToolToken(tool.Name)

	switch {
	case matchesAny(kind, name, "secret", "secrets", "token", "tokens", "credential", "credentials") || hasSensitiveArgSignal(tool.Args):
		return policy.ToolUseSecretExposure, true
	case matchesAny(kind, name, "webfetch", "web_fetch", "websearch", "web_search", "fetch", "http", "network") || argsContainNetworkEgress(tool.Args):
		return policy.ToolUseNetworkEgress, true
	default:
		if matchesAny(kind, name, "patch_apply", "apply_patch", "edit", "write", "multiedit", "multi_edit", "delete", "remove") {
			if len(tool.Args) == 0 || argsTouchProtectedPath(tool.Args) {
				return policy.ToolUseProtectedPathWrite, true
			}
			return "", false
		}
		if matchesAny(kind, name, "shell", "bash", "exec", "command", "run_command", "terminal") {
			command, ok := commandArg(tool.Args)
			if !ok || strings.TrimSpace(command) == "" {
				return policy.ToolUseDestructiveCommand, true
			}
			if commandContainsNetworkEgress(command) {
				return policy.ToolUseNetworkEgress, true
			}
			if commandExposesSecrets(command) {
				return policy.ToolUseSecretExposure, true
			}
			if commandIsDestructive(command) {
				return policy.ToolUseDestructiveCommand, true
			}
			if commandTouchesProtectedPath(command) {
				return policy.ToolUseProtectedPathWrite, true
			}
		}
	}
	return "", false
}

func matchesAny(kind, name string, candidates ...string) bool {
	for _, candidate := range candidates {
		normalized := normalizeToolToken(candidate)
		if kind == normalized || name == normalized {
			return true
		}
	}
	return false
}

func normalizeToolToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

func hasSensitiveArgSignal(args map[string]string) bool {
	for key := range args {
		if toolargs.IsSensitiveKey(key) {
			return true
		}
	}
	return toolargs.HasRedactedValue(args)
}

func argsContainNetworkEgress(args map[string]string) bool {
	for key, value := range args {
		normalizedKey := normalizeToolToken(key)
		normalizedValue := strings.ToLower(strings.TrimSpace(value))
		if strings.Contains(normalizedValue, "https://") || strings.Contains(normalizedValue, "http://") {
			return true
		}
		if strings.Contains(normalizedKey, "url") || strings.Contains(normalizedKey, "uri") || strings.Contains(normalizedKey, "endpoint") {
			if strings.TrimSpace(value) != "" {
				return true
			}
		}
		if commandContainsNetworkEgress(value) {
			return true
		}
	}
	return false
}

func argsTouchProtectedPath(args map[string]string) bool {
	for key, value := range args {
		if !pathLikeArgKey(key) {
			continue
		}
		if isProtectedPath(value) {
			return true
		}
	}
	return false
}

func pathLikeArgKey(key string) bool {
	normalized := normalizeToolToken(key)
	return strings.Contains(normalized, "path") ||
		strings.Contains(normalized, "file") ||
		strings.Contains(normalized, "target")
}
