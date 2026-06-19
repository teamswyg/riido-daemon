package main

import (
	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
)

func baseReq() scheduling.TaskRequirements {
	return scheduling.TaskRequirements{
		Provider:                 "claude",
		AllowExperimentalRuntime: true,
	}
}

func baseCandidate() scheduling.RuntimeCapability {
	return scheduling.RuntimeCapability{
		RuntimeID:             "rt-1",
		Provider:              "claude",
		CapabilityFingerprint: "fp-1",
		Available:             true,
		CompatibilityStatus:   capability.CompatSupported,
		SlotLimit:             2,
		SlotsInUse:            0,
		SupportsSystem:        true,
	}
}

func gateScenario(code string) (
	scheduling.TaskRequirements,
	scheduling.RuntimeCapability,
	bool,
) {
	req := baseReq()
	candidate := baseCandidate()
	switch code {
	case "PROVIDER_MISMATCH":
		req.Provider = "codex"
	case "PROVIDER_UNAVAILABLE":
		candidate.Available = false
	case "COMPATIBILITY_BLOCKED":
		candidate.CompatibilityStatus = capability.CompatBlocked
	case "EXPERIMENTAL_RUNTIME_REQUIRES_OPT_IN":
		req.AllowExperimentalRuntime = false
		candidate.RequiresExperimentalOptIn = true
	case "SLOT_EXHAUSTED":
		candidate.SlotsInUse = candidate.SlotLimit
	case "MISSING_REQUIRED_SURFACE":
		req.RequiredSurfaces = []scheduling.RequiredSurface{scheduling.SurfaceMaxTurns}
	default:
		return req, candidate, false
	}
	return req, candidate, true
}
