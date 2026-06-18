package saasplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *Plane) ResolveToolApproval(ctx context.Context, executionID string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil || !ok {
		return agentbridge.ToolApprovalResolution{}, err
	}
	runtimeID, err := p.runtimeIDForAssignment(ctx, assignment)
	if err != nil {
		return agentbridge.ToolApprovalResolution{}, err
	}
	approvalID := toolApprovalID(tool)
	if err := p.createToolApproval(ctx, assignment, runtimeID, approvalID, tool); err != nil {
		return agentbridge.ToolApprovalResolution{}, err
	}
	return p.waitToolApproval(ctx, assignment, approvalID)
}
