package supervisor

import "context"

func (a *Actor) handleTaskReportEvent(
	ctx context.Context,
	msg *taskEventMsg,
	inFlight map[string]*runningTask,
) {
	if task := inFlight[msg.taskID]; task != nil {
		a.reportTaskEvent(ctx, task, msg.event)
		return
	}
	_ = a.cfg.Reporter.ReportEvent(ctx, msg.taskID, msg.event)
}
