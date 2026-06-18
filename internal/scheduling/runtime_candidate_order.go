package scheduling

import "github.com/teamswyg/riido-contracts/provider/capability"

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
