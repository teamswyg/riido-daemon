package policy

import "fmt"

func validateNativeConfigHookSurfaces(tier TrustTier, surfaces []NativeConfigHookSurface) error {
	seen := map[NativeConfigHookSurface]bool{}
	for _, surface := range surfaces {
		if !isKnownNativeConfigHookSurface(surface) {
			return fmt.Errorf("policy: unknown native config hook surface %q", surface)
		}
		if seen[surface] {
			return fmt.Errorf("policy: duplicate native config hook surface %q", surface)
		}
		seen[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow native config hook surface %q", tier, surface)
		}
	}
	return nil
}

func isKnownNativeConfigHookSurface(surface NativeConfigHookSurface) bool {
	switch surface {
	case NativeConfigHookClaudeCommandAudit:
		return true
	default:
		return false
	}
}
