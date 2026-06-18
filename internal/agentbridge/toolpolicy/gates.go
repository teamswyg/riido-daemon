package toolpolicy

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func PolicyAutoApprover(bundle policy.PolicyBundle, tier policy.TrustTier) agentbridge.AutoApprover {
	return func(tool agentbridge.ToolRef) bool {
		decision, ok := DecisionForTool(bundle, tier, tool)
		return ok && decision.Action == policy.ToolUseActionAllow
	}
}

func PolicyToolStartGate(bundle policy.PolicyBundle, tier policy.TrustTier) agentbridge.ToolStartGate {
	return func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
		decision, ok := DecisionForStartedTool(bundle, tier, tool)
		if !ok || decision.Action == policy.ToolUseActionAllow {
			return agentbridge.ToolStartDecision{}
		}
		return agentbridge.ToolStartDecision{
			Block:  true,
			Code:   decision.Code,
			Reason: decision.Reason,
		}
	}
}

func PolicyToolApprovalGate(bundle policy.PolicyBundle, tier policy.TrustTier) agentbridge.ToolApprovalGate {
	return func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
		decision, ok := DecisionForHeadlessApproval(bundle, tier, tool)
		if !ok || decision.Action == policy.ToolUseActionAllow {
			return agentbridge.ToolStartDecision{}
		}
		reason := strings.TrimSpace(decision.Reason)
		if reason == "" {
			reason = headlessApprovalTimeoutReason
		}
		return agentbridge.ToolStartDecision{
			Block:  true,
			Code:   headlessApprovalTimeoutCode,
			Reason: reason,
		}
	}
}
