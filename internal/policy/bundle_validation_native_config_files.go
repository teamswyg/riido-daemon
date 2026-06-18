package policy

import "fmt"

func validateNativeConfigFileSurfaces(tier TrustTier, surfaces []NativeConfigFileSurface) error {
	seen := map[NativeConfigFileSurface]bool{}
	for _, surface := range surfaces {
		if !isKnownNativeConfigFileSurface(surface) {
			return fmt.Errorf("policy: unknown native config file surface %q", surface)
		}
		if seen[surface] {
			return fmt.Errorf("policy: duplicate native config file surface %q", surface)
		}
		seen[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow native config file surface %q", tier, surface)
		}
	}
	return nil
}

func isKnownNativeConfigFileSurface(surface NativeConfigFileSurface) bool {
	switch surface {
	case NativeConfigFileCodexTaskScopedHome:
		return true
	default:
		return false
	}
}
