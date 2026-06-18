package main

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func daemonPolicySettings(version string, surfaces ...policy.ToolUseSurface) daemonSettings {
	return daemonSettings{PolicyBundleDoc: policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        version,
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {
				AllowedSurfaces: policy.AllowedSurfaceSet{ToolUse: surfaces},
			},
		},
	}}
}
