package saasplane

import (
	"context"
	"log"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *Plane) CompleteTask(ctx context.Context, executionID string, res agentbridge.Result) error {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil || !ok {
		return err
	}
	if res.RecoveryContext != nil {
		c := res.RecoveryContext.Sanitized()
		if journalErr := p.ReportEvent(ctx, executionID, agentbridge.Event{Kind: agentbridge.EventProgress, RecoveryContext: &c}); journalErr != nil {
			log.Printf("riido_daemon event=recovery_journal_delivery_failed state=%q provider=%q", c.State, c.Provider)
		}
	}
	message, err := p.terminalMessage(ctx, executionID, res)
	if err != nil {
		return err
	}
	// Provider completion is not evidence of a delivered answer. Do not run
	// another model turn just to repair an empty response.
	if res.Status == agentbridge.ResultCompleted && strings.TrimSpace(message) == "" && !hasVerifiedResult(res) {
		res.Status = agentbridge.ResultFailed
		res.Error = "The provider finished without an answer or verified result. Retry the request; no automatic retry was started for this empty result."
		message = res.Error
		if res.ExecutionContext != nil {
			execution := *res.ExecutionContext
			execution.DiagnosticCode, execution.NextAction = "provider_failed", "check_provider"
			res.ExecutionContext = &execution
		}
	}
	state, eventType := taskCompletionStateAndEvent(res, message)
	sessionID := res.SessionID
	if res.RecoveryContext != nil && res.RecoveryContext.SwitchCount > 0 {
		sessionID = ""
	}
	request := assignmentcontract.AgentEventRequest{
		AssignmentID:      assignment.ID,
		TaskID:            assignment.TaskID,
		ProviderSessionID: sessionID,
		State:             state,
		EventType:         eventType,
		Message:           message,
		Metadata:          terminalResultMetadata(res, message),
	}
	request.Metadata[metadatakeys.AssignmentEventKey.String()] = terminalEventKey(assignment, executionID, request)
	_, err = p.postAgentEvent(ctx, assignment, request)
	if err != nil {
		return err
	}
	return p.clearCompletedAssignment(ctx, executionID)
}

func taskCompletionStateAndEvent(res agentbridge.Result, message string) (assignmentcontract.AssignmentState, string) {
	if terminalResultNeedsUserInput(res, message) {
		return assignmentcontract.AssignmentRunning, assignmentcontract.EventAssignmentStateUpdated
	}
	return terminalStateAndEvent(res.Status)
}

func (p *Plane) terminalMessage(ctx context.Context, executionID string, res agentbridge.Result) (string, error) {
	message := res.Error
	if message != "" || res.Status != agentbridge.ResultCompleted {
		return firstNonEmpty(message, res.Output), nil
	}
	if strings.TrimSpace(res.Output) != "" {
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
		if st := s.partialBodies[executionID]; st != nil {
			logResponseTiming(executionID, "terminal_ack", st.claimedAt, time.Now())
		}
		delete(s.partialBodies, executionID)
	})
}

func firstNonEmpty(first, second string) string {
	if first != "" {
		return first
	}
	return second
}
