package saasplane

import (
	"context"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *Plane) CompleteTask(ctx context.Context, executionID string, res agentbridge.Result) error {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil || !ok {
		return err
	}
	state, eventType := terminalStateAndEvent(res.Status)
	message, err := p.terminalMessage(ctx, executionID, res)
	if err != nil {
		return err
	}
	_, err = p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{
		AssignmentID:      assignment.ID,
		TaskID:            assignment.TaskID,
		ProviderSessionID: res.SessionID,
		State:             state,
		EventType:         eventType,
		Message:           message,
		Metadata:          terminalResultMetadata(res),
	})
	if err != nil {
		return err
	}
	return p.clearCompletedAssignment(ctx, executionID)
}

func (p *Plane) terminalMessage(ctx context.Context, executionID string, res agentbridge.Result) (string, error) {
	message := res.Error
	if message != "" || res.Status != agentbridge.ResultCompleted {
		return firstNonEmpty(message, res.Output), nil
	}
	if res.Output != "" {
		return res.Output, nil
	}
	err := p.withState(ctx, func(s *planeState) {
		if st := s.partialBodies[executionID]; st != nil {
			message = st.text
		}
	})
	return message, err
}

func (p *Plane) clearCompletedAssignment(ctx context.Context, executionID string) error {
	return p.withState(ctx, func(s *planeState) {
		closeCancelWatcher(s, executionID)
		delete(s.assignmentsByExecution, executionID)
		delete(s.runtimeIDsByExecution, executionID)
		delete(s.partialBodies, executionID)
	})
}

func firstNonEmpty(first, second string) string {
	if first != "" {
		return first
	}
	return second
}
