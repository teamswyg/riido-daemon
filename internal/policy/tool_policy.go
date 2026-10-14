package policy

// EvaluateToolUse implements the current executable subset of security.md G-S3:
// known tiers and explicit bundle allow can proceed; missing allow can only
// continue through a human approval branch; Unknown trust tier always blocks.
func EvaluateToolUse(input ToolUseInput) ToolUseDecision {
	if !isKnownToolUseSurface(input.Surface) {
		return interruptAndBlock("TOOL_USE_UNKNOWN_SURFACE", "tool use surface is not known to the active policy bundle")
	}
	tier := normalizeTrustTier(input.TrustTier)
	switch tier {
	case TrustTierUnknown:
		return interruptAndBlock("TOOL_USE_UNKNOWN_TRUST_TIER", "tool use requires a known runtime trust tier")
	case TrustTierHost, TrustTierIsolatedContainer, TrustTierEphemeralVM, TrustTierCIControlledRunner:
		if input.BundleAllows {
			return ToolUseDecision{Action: ToolUseActionAllow, Code: "ALLOWED", Reason: "tool use surface allowed by active policy bundle"}
		}
		if input.HumanApprovalAvailable {
			return ToolUseDecision{Action: ToolUseActionRequireApproval, Code: "TOOL_USE_REQUIRES_APPROVAL", Reason: "tool use surface is not explicitly allowed by the active policy bundle"}
		}
		return interruptAndBlock("TOOL_USE_NOT_IN_POLICY_BUNDLE", "tool use surface is not explicitly allowed and no approval path is available")
	default:
		return interruptAndBlock("TOOL_USE_UNKNOWN_TRUST_TIER", "tool use requires a known runtime trust tier")
	}
}

func normalizeTrustTier(tier TrustTier) TrustTier {
	if tier == "" {
		return TrustTierUnknown
	}
	return tier
}

func deny(code, reason string) Decision {
	return Decision{Allowed: false, Code: code, Reason: reason}
}

func interruptAndBlock(code, reason string) ToolUseDecision {
	return ToolUseDecision{Action: ToolUseActionInterruptAndBlock, Code: code, Reason: reason}
}
