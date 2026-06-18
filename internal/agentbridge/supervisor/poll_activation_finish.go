package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) finishActivationError(
	ctx context.Context,
	reportCtx context.Context,
	task *runningTask,
	msg *taskActivationMsg,
	inFlight map[string]*runningTask,
) {
	cancelRunningTask(task)
	res := resultForActivationError(msg.err)
	if msg.prepared != nil {
		applyPreparedWorkspace(task, msg.prepared)
		res = a.recordTerminalResult(ctx, task, res)
	}
	_ = a.cfg.Reporter.CompleteTask(reportCtx, task.taskID, res)
	delete(inFlight, task.taskID)
}

func (a *Actor) finishActivationMissingHandle(
	reportCtx context.Context,
	task *runningTask,
	inFlight map[string]*runningTask,
) {
	_ = a.cfg.Reporter.CompleteTask(reportCtx, task.taskID, agentbridge.Result{
		Status: agentbridge.ResultFailed,
		Error:  "supervisor: runtime submit returned no session handle",
	})
	cancelRunningTask(task)
	delete(inFlight, task.taskID)
}

func cancelRunningTask(task *runningTask) {
	if task.cancel == nil {
		return
	}
	task.cancel()
	task.cancel = nil
}
