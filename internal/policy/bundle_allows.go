package policy

import "slices"

func (b PolicyBundle) AllowsUnsafeBypass(tier TrustTier, surface UnsafeBypassSurface) bool {
	tierPolicy, ok := b.TrustTierPolicies[normalizeTrustTier(tier)]
	if !ok {
		return false
	}
	return slices.Contains(tierPolicy.AllowedSurfaces.UnsafeBypass, surface)
}

func (b PolicyBundle) AllowsNativeConfigHook(tier TrustTier, surface NativeConfigHookSurface) bool {
	tierPolicy, ok := b.TrustTierPolicies[normalizeTrustTier(tier)]
	if !ok {
		return false
	}
	return slices.Contains(tierPolicy.AllowedSurfaces.NativeConfigHooks, surface)
}

func (b PolicyBundle) AllowsNativeConfigFile(tier TrustTier, surface NativeConfigFileSurface) bool {
	tierPolicy, ok := b.TrustTierPolicies[normalizeTrustTier(tier)]
	if !ok {
		return false
	}
	return slices.Contains(tierPolicy.AllowedSurfaces.NativeConfigFiles, surface)
}

func (b PolicyBundle) AllowsToolUse(tier TrustTier, surface ToolUseSurface) bool {
	tierPolicy, ok := b.TrustTierPolicies[normalizeTrustTier(tier)]
	if !ok {
		return false
	}
	return slices.Contains(tierPolicy.AllowedSurfaces.ToolUse, surface)
}
