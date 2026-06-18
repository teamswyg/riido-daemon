package saasplane

import (
	"context"

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
		return err
	}
	if ev.Kind == agentbridge.EventTextDelta {
		return p.accumulatePartialBody(ctx, assignment, executionID, ev.Text)
	}
	req, ok := eventRequestFromAgentEvent(assignment, ev)
	if !ok {
		return nil
	}
	_, err = p.postAgentEvent(ctx, assignment, req)
	return err
}
