package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) handleTaskEvent(
	ctx context.Context,
	msg *taskEventMsg,
	inFlight map[string]*runningTask,
	detachedReports *[]detachedReport,
) {
	if task := inFlight[msg.taskID]; task != nil {
		if task.sessionBindingErr != nil {
			return
		}
		if err := a.pinTaskNativeSession(task, msg.event); err != nil {
			task.sessionBindingErr = err
			cancelRunningTask(task)
			a.reportTaskEvent(ctx, task, agentbridge.Event{Kind: agentbridge.EventError, Err: err.Error()})
			return
		}
		if len(msg.originalEvents) == 0 {
			a.appendTaskProviderEvent(ctx, task, msg.event)
		} else {
			for _, ev := range msg.originalEvents {
				a.appendTaskProviderEvent(ctx, task, ev)
			}
		}
		if msg.event.Kind == agentbridge.EventSessionIdentified && task.recovery != nil && task.recovery.SwitchCount > 0 {
			return
		}
		a.reportTaskEvent(ctx, task, msg.event)
		return
	}
	a.reportOrRetainDetached(ctx, detachedReports, detachedEvent(msg.taskID, msg.event))
}
