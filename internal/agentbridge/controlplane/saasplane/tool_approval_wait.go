package saasplane

import (
	"context"
	"fmt"
	"net/url"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *Plane) waitToolApproval(
	ctx context.Context,
	assignment assignmentcontract.Assignment,
	approvalID string,
) (agentbridge.ToolApprovalResolution, error) {
	for {
		var out assignmentcontract.ToolApprovalWaitResponse
		err := p.postJSON(ctx, toolApprovalWaitPath(assignment, approvalID), assignmentcontract.ToolApprovalWaitRequest{
			AssignmentID: assignment.ID,
			WaitMs:       pollWaitMilliseconds(p.cfg.LongPollWait),
		}, &out)
		if err != nil {
			return agentbridge.ToolApprovalResolution{}, err
		}
		resolution, done, err := toolApprovalWaitResolution(ctx, out)
		if err != nil || done {
			return resolution, err
		}
	}
}

func toolApprovalWaitPath(assignment assignmentcontract.Assignment, approvalID string) string {
	return "/v1/agents/" + url.PathEscape(assignment.AgentID) +
		"/tool-approvals/" + url.PathEscape(approvalID) + "/wait"
}

func toolApprovalWaitResolution(
	ctx context.Context,
	out assignmentcontract.ToolApprovalWaitResponse,
) (agentbridge.ToolApprovalResolution, bool, error) {
	switch out.Result.Status {
	case assignmentcontract.ApprovalApproved:
		return agentbridge.ToolApprovalResolution{Approved: true, Reason: toolApprovalDecisionReason(out.Decision)}, true, nil
	case assignmentcontract.ApprovalDenied:
		return agentbridge.ToolApprovalResolution{Reason: toolApprovalDecisionReason(out.Decision)}, true, nil
	case assignmentcontract.ApprovalTimedOut:
		return agentbridge.ToolApprovalResolution{Reason: "tool approval timed out"}, true, nil
	case assignmentcontract.ApprovalPending:
		return agentbridge.ToolApprovalResolution{}, false, ctx.Err()
	default:
		return agentbridge.ToolApprovalResolution{}, true, fmt.Errorf("saasplane: unsupported tool approval status %q", out.Result.Status)
	}
}
