package taskdbplane

import (
	"strings"

	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
)

func runtimeCapabilityForProvider(rec controlplane.RegisteredRuntime, provider string) (scheduling.RuntimeCapability, bool) {
	provider = strings.TrimSpace(provider)
	if rec.RuntimeID == "" || provider == "" {
		return scheduling.RuntimeCapability{}, false
	}
	prefix := "provider." + provider + "."
	available, ok := rec.Capabilities[prefix+"available"]
	if !ok {
		return scheduling.RuntimeCapability{}, false
	}
	return scheduling.RuntimeCapability{
		RuntimeID:                 capability.RuntimeID(rec.RuntimeID),
		Provider:                  capability.ProviderKind(provider),
		CapabilityFingerprint:     capability.CapabilityFingerprint(rec.CapabilityAttributes[prefix+"capability_fingerprint"]),
		SlotLimit:                 rec.SlotLimit,
		SlotsInUse:                rec.SlotsInUse,
		Available:                 available,
		CompatibilityStatus:       capability.CompatibilityStatus(rec.CapabilityAttributes[prefix+"compatibility_status"]),
		RequiresExperimentalOptIn: rec.Capabilities[prefix+"requires_experimental_opt_in"],
		SupportsStreaming:         rec.Capabilities[prefix+"supports_streaming"],
		SupportsResume:            rec.Capabilities[prefix+"supports_resume"],
		SupportsSystem:            rec.Capabilities[prefix+"supports_system"],
		SupportsMaxTurns:          rec.Capabilities[prefix+"supports_max_turns"],
		SupportsMCP:               rec.Capabilities[prefix+"supports_mcp"],
		SupportsToolHooks:         rec.Capabilities[prefix+"supports_tool_hooks"],
		SupportsUsage:             rec.Capabilities[prefix+"supports_usage"],
		SupportsWorktree:          rec.Capabilities[prefix+"supports_worktree"],
	}, true
}
