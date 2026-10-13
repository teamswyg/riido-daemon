package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) handleTaskResult(
	ctx context.Context,
	msg *taskResultMsg,
	inFlight map[string]*runningTask,
	detachedReports *[]detachedReport,
) {
	running := inFlight[msg.taskID]
	reportCtx := ctx
	res := msg.result
	if running != nil {
		res.Usage = running.priorUsage.Add(res.Usage)
		if running.initialStartedAt.IsZero() {
			running.initialStartedAt = res.StartedAt
		} else {
			res.StartedAt = running.initialStartedAt
		}
		if res.Status == agentbridge.ResultCompleted && running.execution != nil {
			a.beginExecutionVerification(running, res)
			return
		}
		res = a.recordTerminalResult(ctx, running, res)
		a.finishTaskWithResult(ctx, inFlight, running, res)
		return
	}
	a.reportOrRetainDetached(reportCtx, detachedReports, detachedResult(msg.taskID, res))
}
