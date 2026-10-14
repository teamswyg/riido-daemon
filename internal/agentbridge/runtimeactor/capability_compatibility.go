package runtimeactor

import providercap "github.com/teamswyg/riido-contracts/provider/capability"

func summarizeCompatibility(
	maturity providercap.ProtocolMaturity,
	blocked []providercap.CompatibilityReason,
	degraded []providercap.CompatibilityReason,
) providercap.CompatibilityStatus {
	switch {
	case len(blocked) > 0:
		return providercap.CompatBlocked
	case maturity == providercap.ProtocolMaturityExperimental || maturity == providercap.ProtocolMaturityDeprecated:
		return providercap.CompatExperimental
	case len(degraded) > 0:
		return providercap.CompatDegraded
	default:
		return providercap.CompatSupported
	}
}
