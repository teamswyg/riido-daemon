package saasplane

import (
	"context"
	"errors"
	"maps"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *Plane) StartTask(ctx context.Context, executionID string) error {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil || !ok {
		return err
	}
	_, err = p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		State:        assignmentcontract.AssignmentReady,
		EventType:    assignmentcontract.EventAssignmentReady,
		Message:      "daemon ready",
	})
	return err
}

func (p *Plane) ReportEvent(ctx context.Context, executionID string, ev agentbridge.Event) error {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil || !ok {
		if err == nil && ev.RecoveryContext != nil {
			return errors.New("recovery assignment authority is unavailable")
		}
		return err
	}
	if ev.Kind == agentbridge.EventTextDelta {
		if assignment.ConfirmationReason != "" {
			if _, err := p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{AssignmentID: assignment.ID, TaskID: assignment.TaskID, State: assignmentcontract.AssignmentRunning, EventType: assignmentcontract.EventAssignmentRunning}); err != nil {
				return err
			}
		}
		return p.accumulatePartialBody(ctx, assignment, executionID, ev.Text)
	}
	req, ok := eventRequestFromAgentEvent(assignment, ev)
	if !ok {
		return nil
	}
	if ev.Kind == agentbridge.EventProgress && ev.RecoveryContext != nil {
		if req.Metadata == nil {
			req.Metadata = map[string]string{}
		}
		maps.Copy(req.Metadata, ev.RecoveryContext.Metadata())
		req.Metadata["lease_token"] = assignment.LeaseToken
		req.Metadata["assignment_event_key"] = recoveryEventKey(assignment, executionID, ev.RecoveryContext.Sanitized())
	}
	if ev.ExecutionContext != nil {
		if req.Metadata == nil {
			req.Metadata = map[string]string{}
		}
		maps.Copy(req.Metadata, ev.ExecutionContext.Metadata())
	}
	if ev.Kind == agentbridge.EventProgress && ev.RecoveryContext != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	response, err := p.postAgentEvent(ctx, assignment, req)
	if err == nil && ev.Kind == agentbridge.EventProgress && ev.RecoveryContext != nil {
		err = validateRecoveryAcknowledgement(assignment, req, response)
	}
	return err
}
