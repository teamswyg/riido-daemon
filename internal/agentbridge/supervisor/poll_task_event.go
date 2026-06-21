package supervisor

import (
	"context"
)

func (a *Actor) handleTaskEvent(
	ctx context.Context,
	msg *taskEventMsg,
	inFlight map[string]*runningTask,
) {
	if task := inFlight[msg.taskID]; task != nil {
		a.appendProviderEvent(ctx, msg.taskID, task.events, msg.event)
		a.reportTaskEvent(ctx, task, msg.event)
		return
	}
	_ = a.cfg.Reporter.ReportEvent(ctx, msg.taskID, msg.event)
}
