package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func decideStartedTool(gate agentbridge.ToolStartGate, tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
	if gate == nil {
		return agentbridge.ToolStartDecision{}
	}
	return gate(tool)
}

func decideApprovalTool(gate agentbridge.ToolApprovalGate, tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
	if gate == nil {
		return agentbridge.ToolStartDecision{}
	}
	return gate(tool)
}

func toolBlockReason(decision agentbridge.ToolStartDecision) string {
	if decision.Code == "" {
		return decision.Reason
	}
	if decision.Reason == "" {
		return decision.Code
	}
	return decision.Code + ": " + decision.Reason
}
