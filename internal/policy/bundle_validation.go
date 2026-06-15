package policy

import (
	"fmt"
)

func validateAllowedSurfaces(tier TrustTier, surfaces AllowedSurfaceSet) error {
	seenUnsafe := map[UnsafeBypassSurface]bool{}
	for _, surface := range surfaces.UnsafeBypass {
		if !isKnownUnsafeBypassSurface(surface) {
			return fmt.Errorf("policy: unknown unsafe bypass surface %q", surface)
		}
		if seenUnsafe[surface] {
			return fmt.Errorf("policy: duplicate unsafe bypass surface %q", surface)
		}
		seenUnsafe[surface] = true
		if tier == TrustTierHost || tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow unsafe bypass surface %q", tier, surface)
		}
	}
	seenHooks := map[NativeConfigHookSurface]bool{}
	for _, surface := range surfaces.NativeConfigHooks {
		if !isKnownNativeConfigHookSurface(surface) {
			return fmt.Errorf("policy: unknown native config hook surface %q", surface)
		}
		if seenHooks[surface] {
			return fmt.Errorf("policy: duplicate native config hook surface %q", surface)
		}
		seenHooks[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow native config hook surface %q", tier, surface)
		}
	}
	seenFiles := map[NativeConfigFileSurface]bool{}
	for _, surface := range surfaces.NativeConfigFiles {
		if !isKnownNativeConfigFileSurface(surface) {
			return fmt.Errorf("policy: unknown native config file surface %q", surface)
		}
		if seenFiles[surface] {
			return fmt.Errorf("policy: duplicate native config file surface %q", surface)
		}
		seenFiles[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow native config file surface %q", tier, surface)
		}
	}
	seenToolUse := map[ToolUseSurface]bool{}
	for _, surface := range surfaces.ToolUse {
		if !isKnownToolUseSurface(surface) {
			return fmt.Errorf("policy: unknown tool use surface %q", surface)
		}
		if seenToolUse[surface] {
			return fmt.Errorf("policy: duplicate tool use surface %q", surface)
		}
		seenToolUse[surface] = true
		if tier == TrustTierUnknown {
			return fmt.Errorf("policy: trust tier %s cannot allow tool use surface %q", tier, surface)
		}
	}
	return nil
}

func isKnownTrustTier(tier TrustTier) bool {
	switch tier {
	case TrustTierHost, TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner, TrustTierUnknown:
		return true
	default:
		return false
	}
}

func isKnownUnsafeBypassSurface(surface UnsafeBypassSurface) bool {
	switch surface {
	case UnsafeBypassClaudePermissions, UnsafeBypassCursorYolo, UnsafeBypassCodexYolo, UnsafeBypassCodexDangerBypass:
		return true
	default:
		return false
	}
}

func isKnownNativeConfigHookSurface(surface NativeConfigHookSurface) bool {
	switch surface {
	case NativeConfigHookClaudeCommandAudit:
		return true
	default:
		return false
	}
}

func isKnownNativeConfigFileSurface(surface NativeConfigFileSurface) bool {
	switch surface {
	case NativeConfigFileCodexTaskScopedHome:
		return true
	default:
		return false
	}
}

func isKnownToolUseSurface(surface ToolUseSurface) bool {
	switch surface {
	case ToolUseNetworkEgress, ToolUseProtectedPathWrite, ToolUseSecretExposure, ToolUseDestructiveCommand:
		return true
	default:
		return false
	}
}
