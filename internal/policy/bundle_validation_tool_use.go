package policy

import "fmt"

func validateToolUseSurfaces(tier TrustTier, surfaces []ToolUseSurface) error {
	seen := map[ToolUseSurface]bool{}
	for _, surface := range surfaces {
		if !isKnownToolUseSurface(surface) {
			return fmt.Errorf("policy: unknown tool use surface %q", surface)
		}
		if seen[surface] {
			return fmt.Errorf("policy: duplicate tool use surface %q", surface)
		}
		seen[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow tool use surface %q", tier, surface)
		}
	}
	return nil
}

func isKnownToolUseSurface(surface ToolUseSurface) bool {
	switch surface {
	case ToolUseNetworkEgress, ToolUseProtectedPathWrite, ToolUseSecretExposure, ToolUseDestructiveCommand:
		return true
	default:
		return false
	}
}
