package runtimeactor

import (
	"time"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type providerCapabilityInput struct {
	runtimeID            string
	provider             string
	res                  agentbridge.DetectResult
	profile              capabilityProfile
	status               providercap.CompatibilityStatus
	requiresExperimental bool
	missingCapabilities  []providercap.CapabilityName
	blockedReasons       []providercap.CompatibilityReason
	degradedReasons      []providercap.CompatibilityReason
	discoveredAt         time.Time
	policyBundleVersion  string
}

func newProviderCapability(input providerCapabilityInput) providercap.ProviderCapability {
	res := input.res
	profile := input.profile
	return providercap.ProviderCapability{
		RuntimeID:                     providercap.RuntimeID(input.runtimeID),
		ProviderKind:                  providercap.ProviderKind(input.provider),
		ProtocolKind:                  profile.protocolKind,
		AdapterID:                     input.provider,
		AdapterVersion:                "riido-agentbridge-adapter.v1",
		ProtocolVersion:               "v1",
		ExecutablePath:                res.Executable,
		Argv0:                         res.Executable,
		DetectedVersion:               res.Version,
		DetectedFingerprint:           detectedFingerprintForExecutable(res.Executable),
		DiscoveredAt:                  input.discoveredAt,
		SupportsStructuredEventStream: res.SupportsStreaming,
		EventStreamFormat:             profile.eventStreamFormat,
		SupportsPartialDeltas:         res.SupportsStreaming,
		SupportsResume:                res.SupportsResume,
		SupportsSessionID:             res.SupportsResume,
		SupportsSessionPin:            providercatalog.IsCodex(input.provider),
		SupportsSystemPrompt:          res.SupportsSystem,
		SupportsMaxTurns:              res.SupportsMaxTurns,
		SupportsToolEvents:            res.SupportsToolHooks,
		SupportsUsageMetrics:          res.SupportsUsage,
		SupportsFileEvents:            profile.supportsFileEvents,
		SupportsPermissionControl:     profile.supportsPermissionControl,
		ExposesUnsafePermissionBypass: profile.exposesUnsafeBypass,
		SupportsApprovalProtocol:      profile.supportsApprovalProtocol,
		SupportsMCP:                   res.SupportsMCP,
		SupportsHookEvents:            res.SupportsToolHooks,
		SupportsWorktree:              profile.supportsWorktree,
		DefaultSandboxMode:            profile.defaultSandboxMode,
		DefaultApprovalPolicy:         profile.defaultApprovalPolicy,
		CompatibilityStatus:           input.status,
		ProtocolMaturity:              profile.protocolMaturity,
		RequiresExperimentalOptIn:     input.requiresExperimental,
		MissingCapabilities:           input.missingCapabilities,
		BlockedReasons:                input.blockedReasons,
		DegradedReasons:               input.degradedReasons,
		Unknown:                       map[string]any{"detect_metadata": copyMetadata(res.Metadata)},
	}
}
