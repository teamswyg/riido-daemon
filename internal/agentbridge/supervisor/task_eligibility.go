package supervisor

import (
	"fmt"

	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
)

func taskEligibility(status runtimeactor.Status, req *bridge.TaskRequest) scheduling.Eligibility {
	capView, ok := findCapability(status.Capabilities, string(req.Provider))
	if !ok {
		return missingProviderEligibility(status, req)
	}
	return scheduling.EvaluateCapability(taskRequirements(req), runtimeCapability(status, capView))
}

func missingProviderEligibility(status runtimeactor.Status, req *bridge.TaskRequest) scheduling.Eligibility {
	return scheduling.Eligibility{
		Eligible:  false,
		RuntimeID: capability.RuntimeID(status.RuntimeID),
		Reasons: []scheduling.IneligibilityReason{{
			Code:   "PROVIDER_NOT_REGISTERED",
			Detail: fmt.Sprintf("provider %q is not registered on runtime %q", req.Provider, status.RuntimeID),
		}},
	}
}

func runtimeCapability(status runtimeactor.Status, capView runtimeactor.Capability) scheduling.RuntimeCapability {
	return scheduling.RuntimeCapability{
		RuntimeID:                 capability.RuntimeID(status.RuntimeID),
		Provider:                  capability.ProviderKind(capView.Provider),
		CapabilityFingerprint:     capability.CapabilityFingerprint(capView.CapabilityFingerprint),
		Available:                 capView.Available,
		CompatibilityStatus:       capability.CompatibilityStatus(capView.CompatibilityStatus),
		RequiresExperimentalOptIn: capView.RequiresExperimentalOptIn,
		SupportsStreaming:         capView.SupportsStreaming,
		SupportsResume:            capView.SupportsResume,
		SupportsSystem:            capView.SupportsSystem,
		SupportsMaxTurns:          capView.SupportsMaxTurns,
		SupportsMCP:               capView.SupportsMCP,
		SupportsToolHooks:         capView.SupportsToolHooks,
		SupportsUsage:             capView.SupportsUsage,
		SupportsWorktree:          capView.SupportsWorktree,
	}
}
