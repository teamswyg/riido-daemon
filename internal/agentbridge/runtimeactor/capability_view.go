package runtimeactor

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func buildRuntimeCapability(runtimeID, provider string, res agentbridge.DetectResult, policyBundleVersion string, discoveredAt time.Time) (Capability, error) {
	domain, err := reconcileProviderCapability(runtimeID, provider, res, policyBundleVersion, discoveredAt)
	if err != nil {
		return Capability{}, err
	}
	return Capability{
		Provider:                  provider,
		Available:                 res.Available,
		Version:                   res.Version,
		Executable:                res.Executable,
		Profile:                   metaProfile(res.Metadata),
		Reason:                    res.Reason,
		HealthStatus:              res.HealthStatus,
		DiagnosticCode:            res.DiagnosticCode,
		ProtocolKind:              string(domain.ProtocolKind),
		AdapterID:                 domain.AdapterID,
		AdapterVersion:            domain.AdapterVersion,
		ProtocolVersion:           domain.ProtocolVersion,
		CompatibilityStatus:       string(domain.CompatibilityStatus),
		CapabilityFingerprint:     string(domain.CapabilityFingerprint),
		DetectedFingerprint:       string(domain.DetectedFingerprint),
		RequiresExperimentalOptIn: domain.RequiresExperimentalOptIn,
		SupportsStreaming:         domain.SupportsStructuredEventStream,
		SupportsResume:            domain.SupportsResume,
		SupportsSystem:            domain.SupportsSystemPrompt,
		SupportsMaxTurns:          domain.SupportsMaxTurns,
		SupportsMCP:               domain.SupportsMCP,
		SupportsToolHooks:         domain.SupportsHookEvents,
		SupportsUsage:             domain.SupportsUsageMetrics,
		SupportsFileEvents:        domain.SupportsFileEvents,
		SupportsWorktree:          domain.SupportsWorktree,
	}, nil
}
