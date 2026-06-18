package saasplane

import (
	"context"
	"net/url"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *Plane) createToolApproval(
	ctx context.Context,
	assignment assignmentcontract.Assignment,
	runtimeID string,
	approvalID string,
	tool agentbridge.ToolRef,
) error {
	var out assignmentcontract.ToolApprovalCreateResponse
	return p.postJSON(ctx, toolApprovalCreatePath(assignment), toolApprovalCreateRequest(
		assignment, runtimeID, approvalID, tool,
	), &out)
}

func toolApprovalCreatePath(assignment assignmentcontract.Assignment) string {
	return "/v1/agents/" + url.PathEscape(assignment.AgentID) + "/tool-approvals"
}

func toolApprovalCreateRequest(
	assignment assignmentcontract.Assignment,
	runtimeID string,
	approvalID string,
	tool agentbridge.ToolRef,
) assignmentcontract.ToolApprovalRequest {
	return assignmentcontract.ToolApprovalRequest{
		ApprovalID:        approvalID,
		AssignmentID:      assignment.ID,
		TaskID:            assignment.TaskID,
		AgentID:           assignment.AgentID,
		RuntimeID:         runtimeID,
		ToolID:            toolApprovalToolID(tool, approvalID),
		ToolKind:          strings.TrimSpace(tool.Kind),
		ToolName:          strings.TrimSpace(tool.Name),
		ProviderRequestID: strings.TrimSpace(tool.ProviderRequestID),
		Reason:            "provider requested tool approval",
	}
}
