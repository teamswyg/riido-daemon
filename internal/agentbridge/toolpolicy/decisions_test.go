package toolpolicy

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestDecisionForToolReturnsRequireApprovalWhenBundleDoesNotAllow(t *testing.T) {
	decision, ok := DecisionForTool(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost, agentbridge.ToolRef{Kind: "patch_apply"})
	if !ok {
		t.Fatal("patch_apply should classify")
	}
	if decision.Action != policy.ToolUseActionRequireApproval || decision.Code != "TOOL_USE_REQUIRES_APPROVAL" {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestDecisionForStartedToolInterruptsWhenBundleDoesNotAllow(t *testing.T) {
	decision, ok := DecisionForStartedTool(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost, agentbridge.ToolRef{Kind: "patch_apply"})
	if !ok {
		t.Fatal("patch_apply should classify")
	}
	if decision.Action != policy.ToolUseActionInterruptAndBlock || decision.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestDecisionForHeadlessApprovalInterruptsWhenBundleDoesNotAllow(t *testing.T) {
	decision, ok := DecisionForHeadlessApproval(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost, agentbridge.ToolRef{Kind: "patch_apply"})
	if !ok {
		t.Fatal("patch_apply should classify")
	}
	if decision.Action != policy.ToolUseActionInterruptAndBlock || decision.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("decision = %+v", decision)
	}
}
