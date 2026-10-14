package toolpolicy

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func DecisionForTool(
	bundle policy.PolicyBundle,
	tier policy.TrustTier,
	tool agentbridge.ToolRef,
) (policy.ToolUseDecision, bool) {
	return decisionForToolSurface(bundle, tier, tool, true)
}

func DecisionForStartedTool(
	bundle policy.PolicyBundle,
	tier policy.TrustTier,
	tool agentbridge.ToolRef,
) (policy.ToolUseDecision, bool) {
	return decisionForToolSurface(bundle, tier, tool, false)
}

func DecisionForHeadlessApproval(
	bundle policy.PolicyBundle,
	tier policy.TrustTier,
	tool agentbridge.ToolRef,
) (policy.ToolUseDecision, bool) {
	return decisionForToolSurface(bundle, tier, tool, false)
}

func decisionForToolSurface(
	bundle policy.PolicyBundle,
	tier policy.TrustTier,
	tool agentbridge.ToolRef,
	humanApprovalAvailable bool,
) (policy.ToolUseDecision, bool) {
	surface, ok := ClassifyToolUseSurface(tool)
	if !ok {
		return policy.ToolUseDecision{}, false
	}
	return policy.EvaluateToolUseWithBundle(bundle, policy.ToolUseInput{
		TrustTier:              tier,
		Surface:                surface,
		HumanApprovalAvailable: humanApprovalAvailable,
	}), true
}
