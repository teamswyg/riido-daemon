package runtimeactor

import (
	"maps"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func profileForProvider(provider string) capabilityProfile {
	providerKind := providercatalog.Normalize(provider)
	switch {
	case providercatalog.IsClaudeFamily(provider):
		return capabilityProfile{
			protocolKind:              providercap.ProtocolClaudeStreamJSON,
			protocolMaturity:          providercap.ProtocolMaturityStable,
			eventStreamFormat:         providercap.EventStreamFormatNDJSON,
			supportsPermissionControl: true,
			exposesUnsafeBypass:       true,
			supportsWorktree:          true,
			defaultSandboxMode:        "unknown",
			defaultApprovalPolicy:     "on-request",
		}
	case providerKind == providercatalog.KindCodex:
		return capabilityProfile{
			protocolKind:              providercap.ProtocolCodexAppServer,
			protocolMaturity:          providercap.ProtocolMaturityExperimental,
			eventStreamFormat:         providercap.EventStreamFormatJSONRPCNotifications,
			supportsPermissionControl: true,
			exposesUnsafeBypass:       true,
			supportsApprovalProtocol:  true,
			supportsWorktree:          true,
			defaultSandboxMode:        "workspace-write",
			defaultApprovalPolicy:     "on-request",
		}
	case providerKind == providercatalog.KindOpenClaw:
		return capabilityProfile{
			protocolKind:          providercap.ProtocolOpenClawAgentJSON,
			protocolMaturity:      providercap.ProtocolMaturityExperimental,
			eventStreamFormat:     providercap.EventStreamFormatNDJSON,
			supportsWorktree:      false,
			defaultSandboxMode:    "unknown",
			defaultApprovalPolicy: "unknown",
		}
	case providerKind == providercatalog.KindCursor:
		return capabilityProfile{
			protocolKind:          providercap.ProtocolCursorAgentStreamJSON,
			protocolMaturity:      providercap.ProtocolMaturityExperimental,
			eventStreamFormat:     providercap.EventStreamFormatNDJSON,
			exposesUnsafeBypass:   true,
			supportsWorktree:      true,
			defaultSandboxMode:    "unknown",
			defaultApprovalPolicy: "unknown",
		}
	default:
		return capabilityProfile{
			protocolKind:          providercap.ProtocolKind(string(providerKind) + "-unknown"),
			protocolMaturity:      providercap.ProtocolMaturityUnknown,
			eventStreamFormat:     providercap.EventStreamFormatUnknown,
			defaultSandboxMode:    "unknown",
			defaultApprovalPolicy: "unknown",
		}
	}
}

func missingCapabilities(res agentbridge.DetectResult) []providercap.CapabilityName {
	checks := []struct {
		name providercap.CapabilityName
		ok   bool
	}{
		{"structured-event-stream", res.SupportsStreaming},
		{"session-resume", res.SupportsResume},
		{"system-prompt", res.SupportsSystem},
		{"max-turns", res.SupportsMaxTurns},
		{"mcp", res.SupportsMCP},
		{"tool-hooks", res.SupportsToolHooks},
		{"usage", res.SupportsUsage},
	}
	out := []providercap.CapabilityName{}
	for _, check := range checks {
		if !check.ok {
			out = append(out, check.name)
		}
	}
	return out
}

func summarizeCompatibility(maturity providercap.ProtocolMaturity, blocked, degraded []providercap.CompatibilityReason) providercap.CompatibilityStatus {
	switch {
	case len(blocked) > 0:
		return providercap.CompatBlocked
	case maturity == providercap.ProtocolMaturityExperimental || maturity == providercap.ProtocolMaturityDeprecated:
		return providercap.CompatExperimental
	case len(degraded) > 0:
		return providercap.CompatDegraded
	default:
		return providercap.CompatSupported
	}
}

func importantSurfaceFlags(c providercap.ProviderCapability) map[string]any {
	return map[string]any{
		"SupportsStructuredEventStream": c.SupportsStructuredEventStream,
		"EventStreamFormat":             c.EventStreamFormat,
		"SupportsPartialDeltas":         c.SupportsPartialDeltas,
		"SupportsResume":                c.SupportsResume,
		"SupportsSessionID":             c.SupportsSessionID,
		"SupportsSessionPin":            c.SupportsSessionPin,
		"SupportsSystemPrompt":          c.SupportsSystemPrompt,
		"SupportsMaxTurns":              c.SupportsMaxTurns,
		"SupportsToolEvents":            c.SupportsToolEvents,
		"SupportsFileEvents":            c.SupportsFileEvents,
		"SupportsUsageMetrics":          c.SupportsUsageMetrics,
		"SupportsPermissionControl":     c.SupportsPermissionControl,
		"ExposesUnsafePermissionBypass": c.ExposesUnsafePermissionBypass,
		"SupportsApprovalProtocol":      c.SupportsApprovalProtocol,
		"SupportsSandbox":               c.SupportsSandbox,
		"SupportsManagedSettings":       c.SupportsManagedSettings,
		"SupportsHookEvents":            c.SupportsHookEvents,
		"SupportsMCP":                   c.SupportsMCP,
		"SupportsWorktree":              c.SupportsWorktree,
		"SupportsJSONSchemaTools":       c.SupportsJSONSchemaTools,
		"ProtocolMaturity":              c.ProtocolMaturity,
		"CompatibilityStatus":           c.CompatibilityStatus,
		"RequiresExperimentalOptIn":     c.RequiresExperimentalOptIn,
	}
}

func copyMetadata(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	maps.Copy(out, in)
	return out
}
