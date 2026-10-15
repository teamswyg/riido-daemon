package saasplane

import (
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func eventRequestFromAgentEvent(assignment assignmentcontract.Assignment, ev agentbridge.Event) (assignmentcontract.AgentEventRequest, bool) {
	req := assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
	}
	if ev.Kind == agentbridge.EventProgress && ev.RecoveryContext != nil {
		req.EventType = assignmentcontract.EventRiidoLog
		req.Message = "Provider recovery: " + ev.RecoveryContext.Sanitized().State
		return req, true
	}
	switch ev.Kind {
	case agentbridge.EventSessionIdentified:
		if ev.SessionID == "" {
			return req, false
		}
		req.EventType = assignmentcontract.EventProviderSessionPinned
		req.ProviderSessionID = ev.SessionID
		req.Message = "provider session pinned"
	case agentbridge.EventProgress:
		req.EventType = assignmentcontract.EventRiidoLog
		req.Message = ev.Text
		req.Metadata = agentbridge.ProgressEventMetadata(ev)
	// NOTE: EventTextDelta is intentionally NOT forwarded. Providers (esp. codex)
	// emit deltas as tiny token/JSON fragments; surfacing each as its own progress
	// line produced incoherent, fragmented output ("code", "\":", "110", ...).
	// The control plane shows structured progress + the final result instead.
	// Coherent live streaming requires accumulating deltas into one evolving
	// message (a separate feature), not one progress line per delta.
	case agentbridge.EventLifecycle:
		switch ev.Phase {
		case agentbridge.StateWaitingToolApproval, agentbridge.StateWaitingUserInput:
			req.EventType = assignmentcontract.EventAssignmentWaitingForUser
			req.State = assignmentcontract.AssignmentRunning
			req.ConfirmationReason = assignmentcontract.ConfirmationApprovalRequired
			if ev.Phase == agentbridge.StateWaitingUserInput {
				req.ConfirmationReason = assignmentcontract.ConfirmationAnswerRequired
			}
		case agentbridge.StateRunning, agentbridge.StateToolRunning:
			req.EventType = assignmentcontract.EventAssignmentRunning
			req.State = assignmentcontract.AssignmentRunning
			req.Message = "provider running"
		default:
			return req, false
		}
	case agentbridge.EventToolApprovalNeeded, agentbridge.EventToolCallStarted, agentbridge.EventToolCallCompleted, agentbridge.EventToolCallFailed:
		req.EventType = assignmentcontract.EventRiidoLog
		switch ev.Kind {
		case agentbridge.EventToolApprovalNeeded:
			if ev.AutomaticallyApproved {
				return req, false
			}
			req.Message = "도구 실행 승인을 기다리고 있습니다: " + ev.Tool.Name
			req.EventType = assignmentcontract.EventAssignmentWaitingForUser
			req.State = assignmentcontract.AssignmentRunning
			req.ConfirmationReason = assignmentcontract.ConfirmationApprovalRequired
			if ev.Tool.Name == "AskUserQuestion" {
				req.ConfirmationReason = assignmentcontract.ConfirmationAnswerRequired
			}
		case agentbridge.EventToolCallStarted:
			req.Message = "도구 실행 중: " + ev.Tool.Name
			req.EventType = assignmentcontract.EventAssignmentRunning
			req.State = assignmentcontract.AssignmentRunning
		case agentbridge.EventToolCallCompleted:
			req.Message = "도구 실행 완료: " + ev.Tool.Name
		case agentbridge.EventToolCallFailed:
			req.Message = "도구 실행 실패: " + ev.Tool.Name
		default:
			return req, false
		}
		req.Metadata = map[string]string{agentbridge.ProgressMessageMetadataKey: "runtime.activity"}
	case agentbridge.EventLog:
		req.EventType = assignmentcontract.EventProviderLog
		req.Message = ev.Text
	case agentbridge.EventWarning:
		req.EventType = assignmentcontract.EventProviderWarning
		req.Message = ev.Text
	case agentbridge.EventError:
		req.EventType = assignmentcontract.EventProviderError
		req.Message = textutil.FirstNonEmptyTrimmed(ev.Err, ev.Text)
	default:
		return req, false
	}
	return req, true
}

func terminalStateAndEvent(status agentbridge.ResultStatus) (assignmentcontract.AssignmentState, string) {
	switch status {
	case agentbridge.ResultCompleted:
		return assignmentcontract.AssignmentCompleted, assignmentcontract.EventAssignmentCompleted
	case agentbridge.ResultCancelled:
		return assignmentcontract.AssignmentCancelled, assignmentcontract.EventAssignmentCancelled
	default:
		return assignmentcontract.AssignmentFailed, assignmentcontract.EventAssignmentFailed
	}
}
