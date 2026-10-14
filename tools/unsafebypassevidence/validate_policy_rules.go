package main

import "github.com/teamswyg/riido-daemon/internal/policy"

func hostUnknownDeny(surface policy.UnsafeBypassSurface) bool {
	for _, tier := range []policy.TrustTier{"", policy.TrustTierUnknown, policy.TrustTierHost} {
		decision := policy.EvaluateUnsafeBypass(policy.UnsafeBypassInput{
			TrustTier:    tier,
			Surface:      surface,
			BundleAllows: true,
		})
		if decision.Allowed {
			return false
		}
	}
	return true
}

func isolatedRequiresBundle(surface policy.UnsafeBypassSurface) bool {
	denied := policy.EvaluateUnsafeBypass(policy.UnsafeBypassInput{
		TrustTier:    policy.TrustTierIsolatedContainer,
		Surface:      surface,
		BundleAllows: false,
	})
	allowed := policy.EvaluateUnsafeBypass(policy.UnsafeBypassInput{
		TrustTier:    policy.TrustTierEphemeralVM,
		Surface:      surface,
		BundleAllows: true,
	})
	return !denied.Allowed && allowed.Allowed
}
