package policy

func validateAllowedSurfaces(tier TrustTier, surfaces AllowedSurfaceSet) error {
	if err := validateUnsafeBypassSurfaces(tier, surfaces.UnsafeBypass); err != nil {
		return err
	}
	if err := validateNativeConfigHookSurfaces(tier, surfaces.NativeConfigHooks); err != nil {
		return err
	}
	if err := validateNativeConfigFileSurfaces(tier, surfaces.NativeConfigFiles); err != nil {
		return err
	}
	return validateToolUseSurfaces(tier, surfaces.ToolUse)
}
