package scheduling

import (
	"sort"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// RuntimeSelection is the deterministic result of selecting one runtime
// from a pool for a task.
type RuntimeSelection struct {
	Runtime     RuntimeCapability
	Eligibility Eligibility
	Rejected    []Eligibility
}

// SelectRuntime evaluates a runtime pool and returns the best eligible
// candidate. It is pure C5 logic: no locks, leases, persistence, or
// provider execution side effects happen here.
func SelectRuntime(req TaskRequirements, candidates []RuntimeCapability) (RuntimeSelection, bool) {
	eligible := []RuntimeSelection{}
	rejected := []Eligibility{}
	for _, candidate := range candidates {
		evaluation := EvaluateCapability(req, candidate)
		if evaluation.Eligible {
			eligible = append(eligible, RuntimeSelection{
				Runtime:     candidate,
				Eligibility: evaluation,
			})
			continue
		}
		rejected = append(rejected, evaluation)
	}
	sort.Slice(rejected, func(i, j int) bool {
		return eligibilityLess(rejected[i], rejected[j])
	})
	if len(eligible) == 0 {
		return RuntimeSelection{Rejected: rejected}, false
	}
	sort.Slice(eligible, func(i, j int) bool {
		return runtimeCandidateLess(eligible[i].Runtime, eligible[j].Runtime)
	})
	selected := eligible[0]
	selected.Rejected = rejected
	return selected, true
}

func runtimeCandidateLess(left, right RuntimeCapability) bool {
	leftStatus := compatibilityRank(left.CompatibilityStatus)
	rightStatus := compatibilityRank(right.CompatibilityStatus)
	if leftStatus != rightStatus {
		return leftStatus < rightStatus
	}
	leftHeadroom := slotHeadroom(left)
	rightHeadroom := slotHeadroom(right)
	if leftHeadroom != rightHeadroom {
		return leftHeadroom > rightHeadroom
	}
	if left.SlotsInUse != right.SlotsInUse {
		return left.SlotsInUse < right.SlotsInUse
	}
	if left.RuntimeID != right.RuntimeID {
		return left.RuntimeID < right.RuntimeID
	}
	if left.CapabilityFingerprint != right.CapabilityFingerprint {
		return left.CapabilityFingerprint < right.CapabilityFingerprint
	}
	return left.Provider < right.Provider
}

func compatibilityRank(status capability.CompatibilityStatus) int {
	switch status {
	case capability.CompatSupported:
		return 0
	case capability.CompatDegraded:
		return 1
	case capability.CompatExperimental:
		return 2
	case capability.CompatBlocked:
		return 3
	default:
		return 4
	}
}

func slotHeadroom(candidate RuntimeCapability) int {
	if candidate.SlotLimit <= 0 {
		return 1 << 30
	}
	headroom := candidate.SlotLimit - candidate.SlotsInUse
	if headroom < 0 {
		return 0
	}
	return headroom
}

func eligibilityLess(left, right Eligibility) bool {
	if left.RuntimeID != right.RuntimeID {
		return left.RuntimeID < right.RuntimeID
	}
	if left.CapabilityFingerprint != right.CapabilityFingerprint {
		return left.CapabilityFingerprint < right.CapabilityFingerprint
	}
	return left.Summary() < right.Summary()
}
