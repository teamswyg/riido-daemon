package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) finishActivationError(
	ctx context.Context,
	task *runningTask,
	msg *taskActivationMsg,
	inFlight map[string]*runningTask,
) {
	res := resultForActivationError(msg.err)
	if msg.prepared != nil {
		applyPreparedWorkspace(task, msg.prepared)
		res = a.recordTerminalResult(ctx, task, res)
	}
	a.finishTaskWithResult(ctx, inFlight, task, res)
}

func (a *Actor) finishActivationMissingHandle(
	ctx context.Context,
	task *runningTask,
	inFlight map[string]*runningTask,
) {
	a.finishTaskWithResult(ctx, inFlight, task, agentbridge.Result{
		Status: agentbridge.ResultFailed,
		Error:  "supervisor: runtime submit returned no session handle",
	})
}

func cancelRunningTask(task *runningTask) {
	if task.cancel == nil {
		return
	}
	task.cancel()
	task.cancel = nil
}
