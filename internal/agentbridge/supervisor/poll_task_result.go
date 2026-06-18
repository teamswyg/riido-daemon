package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
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
		reportCtx = controlplane.ContextWithTaskReport(ctx, running.report)
		res = a.recordTerminalResult(ctx, running, msg.result)
		cancelRunningTask(running)
	}
	_ = a.cfg.Reporter.CompleteTask(reportCtx, msg.taskID, res)
	delete(inFlight, msg.taskID)
}
