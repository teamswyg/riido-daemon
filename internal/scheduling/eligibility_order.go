package scheduling

func eligibilityLess(left, right Eligibility) bool {
	if left.RuntimeID != right.RuntimeID {
		return left.RuntimeID < right.RuntimeID
	}
	if left.CapabilityFingerprint != right.CapabilityFingerprint {
		return left.CapabilityFingerprint < right.CapabilityFingerprint
	}
	return left.Summary() < right.Summary()
}
