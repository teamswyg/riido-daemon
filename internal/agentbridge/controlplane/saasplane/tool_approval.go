package saasplane

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
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

func (p *Plane) createToolApproval(ctx context.Context, assignment assignmentcontract.Assignment, runtimeID, approvalID string, tool agentbridge.ToolRef) error {
	var out assignmentcontract.ToolApprovalCreateResponse
	return p.postJSON(ctx, "/v1/agents/"+url.PathEscape(assignment.AgentID)+"/tool-approvals", assignmentcontract.ToolApprovalRequest{
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
	}, &out)
}

func (p *Plane) waitToolApproval(ctx context.Context, assignment assignmentcontract.Assignment, approvalID string) (agentbridge.ToolApprovalResolution, error) {
	for {
		var out assignmentcontract.ToolApprovalWaitResponse
		err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(assignment.AgentID)+"/tool-approvals/"+url.PathEscape(approvalID)+"/wait", assignmentcontract.ToolApprovalWaitRequest{
			AssignmentID: assignment.ID,
			WaitMs:       pollWaitMilliseconds(p.cfg.LongPollWait),
		}, &out)
		if err != nil {
			return agentbridge.ToolApprovalResolution{}, err
		}
		switch out.Result.Status {
		case assignmentcontract.ApprovalApproved:
			return agentbridge.ToolApprovalResolution{Approved: true, Reason: toolApprovalDecisionReason(out.Decision)}, nil
		case assignmentcontract.ApprovalDenied:
			return agentbridge.ToolApprovalResolution{Reason: toolApprovalDecisionReason(out.Decision)}, nil
		case assignmentcontract.ApprovalTimedOut:
			return agentbridge.ToolApprovalResolution{Reason: "tool approval timed out"}, nil
		case assignmentcontract.ApprovalPending:
			if err := ctx.Err(); err != nil {
				return agentbridge.ToolApprovalResolution{}, err
			}
		default:
			return agentbridge.ToolApprovalResolution{}, fmt.Errorf("saasplane: unsupported tool approval status %q", out.Result.Status)
		}
	}
}

func toolApprovalID(tool agentbridge.ToolRef) string {
	if id := strings.TrimSpace(tool.ID); id != "" {
		return id
	}
	if id := strings.TrimSpace(tool.ProviderRequestID); id != "" {
		return id
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(tool.Kind) + "\x00" + strings.TrimSpace(tool.Name) + "\x00" + time.Now().UTC().Format(time.RFC3339Nano)))
	return "approval-" + hex.EncodeToString(sum[:8])
}

func toolApprovalToolID(tool agentbridge.ToolRef, fallback string) string {
	if id := strings.TrimSpace(tool.ID); id != "" {
		return id
	}
	if id := strings.TrimSpace(tool.ProviderRequestID); id != "" {
		return id
	}
	return fallback
}

func toolApprovalDecisionReason(decision *assignmentcontract.ToolApprovalDecision) string {
	if decision == nil {
		return ""
	}
	return strings.TrimSpace(decision.Reason)
}
