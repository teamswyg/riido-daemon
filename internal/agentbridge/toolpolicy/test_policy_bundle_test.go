package toolpolicy

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func testPolicyBundle(surfaces policy.AllowedSurfaceSet) policy.PolicyBundle {
	return policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.toolpolicy-test.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {
				AllowedSurfaces: surfaces,
			},
		},
	}
}
