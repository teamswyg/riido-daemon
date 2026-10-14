package supervisor

import (
	"context"
)

func (a *Actor) handleTaskEvent(
	ctx context.Context,
	msg *taskEventMsg,
	inFlight map[string]*runningTask,
	detachedReports *[]detachedReport,
) {
	if task := inFlight[msg.taskID]; task != nil {
		a.appendTaskProviderEvent(ctx, task, msg.event)
		a.reportTaskEvent(ctx, task, msg.event)
		return
	}
	a.reportOrRetainDetached(ctx, detachedReports, detachedEvent(msg.taskID, msg.event))
}
