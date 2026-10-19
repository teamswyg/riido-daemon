package supervisor

import (
	"context"
	"log"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Only a cancellation handled for this claim changes a not-yet-attempted result.
// The context alone is insufficient: finishTaskWithResult cancels it internally.
// A previously attempted result may already be committed with a lost ACK.
func resultAfterClaimCancellation(task *runningTask, res agentbridge.Result) agentbridge.Result {
	if task != nil && !task.terminalDeliveryAttempted && task.cancelCause != nil && res.Status != agentbridge.ResultCancelled {
		res.Status = agentbridge.ResultCancelled
		res.Error = "Execution cancelled before terminal delivery"
	}
	return res
}

func claimDeliveryCancelled(task *runningTask) bool {
	return task != nil && (task.cancelCause != nil || task.terminalResult != nil && task.terminalResult.Status == agentbridge.ResultCancelled)
}

func obsoleteAfterCancellation(ev agentbridge.Event) bool {
	if ev.Kind == agentbridge.EventToolApprovalNeeded {
		return true
	}
	if ev.Kind != agentbridge.EventLifecycle {
		return false
	}
	switch ev.Phase {
	case agentbridge.StateRunning, agentbridge.StateToolRunning, agentbridge.StateWaitingUserInput, agentbridge.StateWaitingToolApproval:
		return true
	default:
		return false
	}
}

// Obsolete intermediate notifications cannot hold a cancelled assignment's
// terminal report behind a permanent server state-transition rejection. Keep
// bounded evidence without discarding session pins or other delivery records.
func (a *Actor) supersedeCancelledNotifications(ctx context.Context, task *runningTask) {
	if !claimDeliveryCancelled(task) {
		return
	}
	pending := task.pendingEvents
	task.pendingEvents = nil
	for _, ev := range pending {
		if obsoleteAfterCancellation(ev) {
			a.recordSupersededNotification(ctx, task, ev)
		} else {
			task.pendingEvents = append(task.pendingEvents, ev)
		}
	}
}

func (a *Actor) recordSupersededNotification(ctx context.Context, task *runningTask, ev agentbridge.Event) {
	if len(task.supersededEvents) >= maxPendingReportEvents {
		return
	}
	task.supersededEvents = append(task.supersededEvents, ev)
	// The original provider event was already journaled. Record only bounded
	// classification in the additional diagnostic, never tool arguments/content.
	kind, phase := supersededNotificationClassification(ev)
	a.appendTaskProviderEvent(ctx, task, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "Intermediate notification superseded after claim cancellation: " + kind + "/" + phase})
	log.Printf("riido_daemon event=cancelled_notification_superseded kind=%q phase=%q", kind, phase)
}

func supersededNotificationClassification(ev agentbridge.Event) (string, string) {
	kind := "unknown"
	switch ev.Kind {
	case agentbridge.EventLifecycle, agentbridge.EventToolApprovalNeeded:
		kind = string(ev.Kind)
	default:
	}
	phase := "unknown"
	switch ev.Phase {
	case "":
		phase = "none"
	case agentbridge.StateRunning, agentbridge.StateToolRunning, agentbridge.StateWaitingUserInput, agentbridge.StateWaitingToolApproval:
		phase = string(ev.Phase)
	default:
	}
	return kind, phase
}

// A buffered result held behind intermediate notifications has never crossed the
// delivery boundary. An explicit cancellation may still supersede it; append a
// correction and update the keep-in-place archive instead of erasing evidence.
func (a *Actor) reconcileUnattemptedCancellation(ctx context.Context, task *runningTask) {
	if task.terminalResult == nil || task.terminalDeliveryAttempted || task.cancelCause == nil || task.terminalResult.Status == agentbridge.ResultCancelled {
		return
	}
	res := resultAfterClaimCancellation(task, *task.terminalResult)
	res = a.recordTerminalResult(ctx, task, res)
	res = terminalRecoveryContext(a, task, res)
	res.ExecutionContext = resultExecutionContext(task, res)
	task.terminalResult = &res
}
