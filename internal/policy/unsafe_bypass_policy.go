package policy

// UnsafeBypassSurface is one concrete provider surface covered by
// security.md §5.
type UnsafeBypassSurface string

const (
	UnsafeBypassClaudePermissions UnsafeBypassSurface = "claude:bypassPermissions"
	UnsafeBypassCursorYolo        UnsafeBypassSurface = "cursor:--yolo"
	UnsafeBypassCodexYolo         UnsafeBypassSurface = "codex:--yolo"
	UnsafeBypassCodexDangerBypass UnsafeBypassSurface = "codex:--dangerously-bypass-approvals-and-sandbox"
)

// UnsafeBypassInput is the policy question: may this runtime activate a
// provider unsafe-bypass surface?
type UnsafeBypassInput struct {
	TrustTier     TrustTier
	Surface       UnsafeBypassSurface
	BundleAllows  bool
	PolicyVersion string
}

// EvaluateUnsafeBypass implements security.md §5:
// Host/Unknown always deny; isolated tiers require an explicit bundle allow.
func EvaluateUnsafeBypass(input UnsafeBypassInput) Decision {
	tier := normalizeTrustTier(input.TrustTier)
	switch tier {
	case TrustTierHost:
		return deny("UNSAFE_BYPASS_ON_HOST", "unsafe bypass is forbidden on Host trust tier")
	case TrustTierUnknown:
		return deny("UNSAFE_BYPASS_UNKNOWN_TRUST_TIER", "unsafe bypass requires a verified isolated trust tier")
	case TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner:
		if !input.BundleAllows {
			return deny("UNSAFE_BYPASS_NOT_IN_POLICY_BUNDLE", "unsafe bypass is not explicitly allowed by the active policy bundle")
		}
		return Decision{Allowed: true, Code: "ALLOWED", Reason: "unsafe bypass allowed by isolated trust tier and active policy bundle"}
	default:
		return deny("UNSAFE_BYPASS_UNKNOWN_TRUST_TIER", "unsafe bypass requires a known trust tier")
	}
}
