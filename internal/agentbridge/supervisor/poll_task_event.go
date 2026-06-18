package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (a *Actor) handleTaskEvent(
	ctx context.Context,
	msg *taskEventMsg,
	inFlight map[string]*runningTask,
) {
	reportCtx := ctx
	if task := inFlight[msg.taskID]; task != nil {
		reportCtx = controlplane.ContextWithTaskReport(ctx, task.report)
		a.appendProviderEvent(ctx, msg.taskID, task.events, msg.event)
	}
	_ = a.cfg.Reporter.ReportEvent(reportCtx, msg.taskID, msg.event)
}
