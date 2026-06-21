package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) cancelInFlightTasks(ctx context.Context, inFlight map[string]*runningTask, finishedAt time.Time) {
	for _, task := range inFlight {
		a.cancelInFlightTask(ctx, inFlight, task, finishedAt)
	}
	a.drainShutdownReports(ctx, inFlight)
}

func (a *Actor) cancelInFlightTask(
	ctx context.Context,
	inFlight map[string]*runningTask,
	task *runningTask,
	finishedAt time.Time,
) {
	if task.handle != nil {
		_ = task.runtime.Cancel(ctx, task.taskID, ErrStopped.Error())
	}
	if task.cancel != nil {
		task.cancel()
		task.cancel = nil
	}
	res := a.recordTerminalResult(ctx, task, agentbridge.Result{
		Status:     agentbridge.ResultCancelled,
		Error:      ErrStopped.Error(),
		FinishedAt: finishedAt,
	})
	a.finishTaskWithResult(ctx, inFlight, task, res)
}
