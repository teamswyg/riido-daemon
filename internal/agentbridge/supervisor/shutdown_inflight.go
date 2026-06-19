package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (a *Actor) cancelInFlightTasks(ctx context.Context, inFlight map[string]*runningTask, finishedAt time.Time) {
	for taskID, task := range inFlight {
		a.cancelInFlightTask(ctx, task, finishedAt)
		delete(inFlight, taskID)
	}
}

func (a *Actor) cancelInFlightTask(ctx context.Context, task *runningTask, finishedAt time.Time) {
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
	_ = a.cfg.Reporter.CompleteTask(controlplane.ContextWithTaskReport(ctx, task.report), task.taskID, res)
}
