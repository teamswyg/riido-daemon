package policy

func isKnownTrustTier(tier TrustTier) bool {
	switch tier {
	case TrustTierHost, TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner, TrustTierUnknown:
		return true
	default:
		return false
	}
}
