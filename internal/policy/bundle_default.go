package policy

import "time"

func DefaultLocalPolicyBundle() PolicyBundle {
	return PolicyBundle{
		SchemaVersion:  BundleSchemaVersion,
		Version:        DefaultLocalPolicyBundleVersion,
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[TrustTier]TrustTierPolicy{
			TrustTierHost: {
				AllowedSurfaces: AllowedSurfaceSet{
					NativeConfigHooks: []NativeConfigHookSurface{
						NativeConfigHookClaudeCommandAudit,
					},
				},
			},
		},
	}
}
