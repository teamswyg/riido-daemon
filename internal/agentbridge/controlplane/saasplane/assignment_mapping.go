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
	switch ev.Kind {
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
		if ev.Phase == agentbridge.StateRunning {
			req.EventType = assignmentcontract.EventAssignmentRunning
			req.State = assignmentcontract.AssignmentRunning
			req.Message = "provider running"
		} else {
			return req, false
		}
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
