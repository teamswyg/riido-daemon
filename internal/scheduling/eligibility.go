package scheduling

import (
	"fmt"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// IneligibilityReason is one reason a runtime cannot execute a task.
type IneligibilityReason struct {
	Code    string
	Surface RequiredSurface
	Detail  string
}

// Eligibility is the deterministic result of comparing one task's
// requirements against one runtime capability snapshot.
type Eligibility struct {
	Eligible              bool
	RuntimeID             capability.RuntimeID
	CapabilityFingerprint capability.CapabilityFingerprint
	Reasons               []IneligibilityReason
}

// EvaluateCapability applies the C5 pre-submit scheduling gate.
func EvaluateCapability(req TaskRequirements, candidate RuntimeCapability) Eligibility {
	out := newEligibleCapability(candidate)
	if req.Provider != "" && req.Provider != candidate.Provider {
		out.add("PROVIDER_MISMATCH", "", fmt.Sprintf("task provider %q does not match runtime provider %q", req.Provider, candidate.Provider))
	}
	if !candidate.Available {
		out.add("PROVIDER_UNAVAILABLE", "", fmt.Sprintf("provider %q is unavailable", candidate.Provider))
	}
	if candidate.compatibilityBlocked() {
		out.add("COMPATIBILITY_BLOCKED", "", fmt.Sprintf("provider %q compatibility is blocked", candidate.Provider))
	}
	if candidate.RequiresExperimentalOptIn && !req.AllowExperimentalRuntime {
		out.add("EXPERIMENTAL_RUNTIME_REQUIRES_OPT_IN", "", fmt.Sprintf("provider %q requires allow_experimental_runtime", candidate.Provider))
	}
	if candidate.slotsExhausted() {
		out.add("SLOT_EXHAUSTED", "", fmt.Sprintf("runtime %q has no free execution slots", candidate.RuntimeID))
	}
	addSurfaceReasons(&out, req, candidate)
	return out
}

func newEligibleCapability(candidate RuntimeCapability) Eligibility {
	return Eligibility{
		Eligible:              true,
		RuntimeID:             candidate.RuntimeID,
		CapabilityFingerprint: candidate.CapabilityFingerprint,
	}
}
