package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (a *Actor) finishTaskWithResult(
	ctx context.Context,
	inFlight map[string]*runningTask,
	task *runningTask,
	res agentbridge.Result,
) bool {
	cancelRunningTask(task)
	task.terminalResult = &res
	return a.flushTerminalReport(ctx, inFlight, task)
}

func (a *Actor) flushTerminalReport(
	ctx context.Context,
	inFlight map[string]*runningTask,
	task *runningTask,
) bool {
	if task == nil || task.terminalResult == nil {
		return false
	}
	if len(task.pendingEvents) > 0 {
		return false
	}
	reportCtx := controlplane.ContextWithTaskReport(ctx, task.report)
	if err := a.cfg.Reporter.CompleteTask(reportCtx, task.taskID, *task.terminalResult); err != nil {
		return false
	}
	delete(inFlight, task.taskID)
	return true
}

func (a *Actor) retryTerminalReports(ctx context.Context, inFlight map[string]*runningTask) bool {
	reported := false
	for _, task := range inFlight {
		if task.terminalResult == nil {
			continue
		}
		if a.flushTerminalReport(ctx, inFlight, task) {
			reported = true
		}
	}
	return reported
}
