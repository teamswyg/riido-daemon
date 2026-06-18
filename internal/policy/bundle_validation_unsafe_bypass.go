package policy

import "fmt"

func validateUnsafeBypassSurfaces(tier TrustTier, surfaces []UnsafeBypassSurface) error {
	seen := map[UnsafeBypassSurface]bool{}
	for _, surface := range surfaces {
		if !isKnownUnsafeBypassSurface(surface) {
			return fmt.Errorf("policy: unknown unsafe bypass surface %q", surface)
		}
		if seen[surface] {
			return fmt.Errorf("policy: duplicate unsafe bypass surface %q", surface)
		}
		seen[surface] = true
		if tier == TrustTierHost || tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow unsafe bypass surface %q", tier, surface)
		}
	}
	return nil
}

func isKnownUnsafeBypassSurface(surface UnsafeBypassSurface) bool {
	switch surface {
	case UnsafeBypassClaudePermissions, UnsafeBypassCursorYolo, UnsafeBypassCodexYolo, UnsafeBypassCodexDangerBypass:
		return true
	default:
		return false
	}
}
