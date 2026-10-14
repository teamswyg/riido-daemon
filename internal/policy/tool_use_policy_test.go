package policy

import "testing"

func TestEvaluateToolUseBranchesByPolicyAndApproval(t *testing.T) {
	allowed := EvaluateToolUse(ToolUseInput{
		TrustTier:    TrustTierHost,
		Surface:      ToolUseNetworkEgress,
		BundleAllows: true,
	})
	if allowed.Action != ToolUseActionAllow {
		t.Fatalf("explicit tool use allow = %+v, want allow", allowed)
	}

	approval := EvaluateToolUse(ToolUseInput{
		TrustTier:              TrustTierHost,
		Surface:                ToolUseProtectedPathWrite,
		HumanApprovalAvailable: true,
	})
	if approval.Action != ToolUseActionRequireApproval || approval.Code != "TOOL_USE_REQUIRES_APPROVAL" {
		t.Fatalf("missing tool use allow with approval = %+v, want require approval", approval)
	}

	blocked := EvaluateToolUse(ToolUseInput{
		TrustTier: TrustTierHost,
		Surface:   ToolUseSecretExposure,
	})
	if blocked.Action != ToolUseActionInterruptAndBlock || blocked.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("missing tool use allow without approval = %+v, want interrupt-and-block", blocked)
	}
}

func TestEvaluateToolUseBlocksUnknownTierAndSurface(t *testing.T) {
	for _, tier := range []TrustTier{"", TrustTierUnknown, TrustTier("MoonBase")} {
		got := EvaluateToolUse(ToolUseInput{
			TrustTier:    tier,
			Surface:      ToolUseNetworkEgress,
			BundleAllows: true,
		})
		if got.Action != ToolUseActionInterruptAndBlock || got.Code != "TOOL_USE_UNKNOWN_TRUST_TIER" {
			t.Fatalf("tier %q tool use decision = %+v, want unknown tier block", tier, got)
		}
	}

	got := EvaluateToolUse(ToolUseInput{
		TrustTier:    TrustTierHost,
		Surface:      ToolUseSurface("tool:ghost"),
		BundleAllows: true,
	})
	if got.Action != ToolUseActionInterruptAndBlock || got.Code != "TOOL_USE_UNKNOWN_SURFACE" {
		t.Fatalf("unknown surface tool use decision = %+v, want unknown surface block", got)
	}
}
