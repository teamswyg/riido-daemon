package supervisor

import (
	"context"
)

func (a *Actor) handleTaskResult(
	ctx context.Context,
	msg *taskResultMsg,
	inFlight map[string]*runningTask,
) {
	running := inFlight[msg.taskID]
	reportCtx := ctx
	res := msg.result
	if running != nil {
		res = a.recordTerminalResult(ctx, running, msg.result)
		a.finishTaskWithResult(ctx, inFlight, running, res)
		return
	}
	_ = a.cfg.Reporter.CompleteTask(reportCtx, msg.taskID, res)
}
