package saasplane

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
)

func (p *Plane) CompleteTask(ctx context.Context, executionID string, res agentbridge.Result) error {
	binding, err := p.claimBinding(ctx, executionID)
	if err != nil {
		return err
	}
	if binding != nil {
		binding.mu.Lock()
		defer binding.mu.Unlock()
		if binding.terminal != nil {
			return p.deliverClaimTerminal(ctx, executionID, binding, *binding.terminal)
		}
	}
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if binding != nil {
		assignment, ok, err = binding.assignment, true, nil
	}
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
	if binding != nil {
		binding.terminal = &request
		proof, known := executionproof.ProofFromContext(ctx)
		observed, complete := executionproof.ScopeFromContext(ctx).Tracker().Seal()
		binding.proof, binding.known = proof, known && complete && proof == observed && assignmentcontract.IsTerminal(request.State)
		return p.deliverClaimTerminal(ctx, executionID, binding, request)
	}
	if _, err = p.postAgentEvent(ctx, assignment, request); err != nil {
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
		if !currentClaimScope(ctx, s, executionID) {
			return
		}
		if st := s.partialBodies[executionID]; st != nil {
			message = st.text
		}
	})
	return message, err
}

func (p *Plane) clearCompletedAssignment(ctx context.Context, executionID string) error {
	return p.withState(ctx, func(s *planeState) {
		if !currentClaimScope(ctx, s, executionID) {
			return
		}
		closeCancelWatcher(s, executionID)
		delete(s.executionScopes, executionID)
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

// A known observation participates in the existing terminal retry. The same
// immutable event/key is repeated only while its ACK is unknown. Once that ACK
// is confirmed, retry the receipt directly even after a new epoch/binding.
func (p *Plane) deliverClaimTerminal(ctx context.Context, executionID string, binding *executionClaimBinding, request assignmentcontract.AgentEventRequest) error {
	if !binding.terminalACK {
		if _, err := p.postAgentEvent(ctx, binding.assignment, request); err != nil {
			return err
		}
		binding.terminalACK = true
	}
	if binding.known {
		if err := p.postExecutionTermination(ctx, binding, binding.proof); err != nil {
			if !errors.Is(err, errTerminationUnsupported) {
				return err
			}
			// Older control planes have no receipt route. Business terminal ACK is
			// still authoritative, but no termination receipt/permission is inferred.
			log.Print("riido_daemon event=execution_termination_unsupported observation=unknown")
		}
	}
	return p.clearCompletedAssignment(ctx, executionID)
}
