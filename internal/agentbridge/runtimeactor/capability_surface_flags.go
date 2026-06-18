package runtimeactor

import providercap "github.com/teamswyg/riido-contracts/provider/capability"

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
