package supervisor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

const maxPendingReportEvents = 16

func shouldRetainReportEvent(ev agentbridge.Event) bool {
	switch ev.Kind {
	case agentbridge.EventLifecycle,
		agentbridge.EventSessionIdentified,
		agentbridge.EventToolApprovalNeeded,
		agentbridge.EventCancellation,
		agentbridge.EventTimeout,
		agentbridge.EventProcessExit:
		return true
	default:
		return false
	}
}
