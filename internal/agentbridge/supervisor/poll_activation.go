package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (a *Actor) handleTaskActivation(
	ctx context.Context,
	msg *taskActivationMsg,
	inFlight map[string]*runningTask,
) bool {
	task := inFlight[msg.taskID]
	if task == nil {
		return false
	}
	reportCtx := controlplane.ContextWithTaskReport(ctx, task.report)
	if msg.err != nil {
		a.finishActivationError(ctx, task, msg, inFlight)
		return true
	}
	if msg.handle == nil {
		a.finishActivationMissingHandle(ctx, task, inFlight)
		return true
	}
	task.handle = msg.handle
	applyPreparedWorkspace(task, msg.prepared)
	if task.cancelCause != nil || task.ctx.Err() != nil {
		a.cancelActivatedTask(ctx, task)
		forwardActivatedSession(a, task)
		return false
	}
	_ = a.cfg.Reporter.ReportEvent(reportCtx, task.taskID, agentbridge.Event{
		Kind:  agentbridge.EventLifecycle,
		Phase: agentbridge.StateRunning,
	})
	forwardActivatedSession(a, task)
	return false
}

func applyPreparedWorkspace(task *runningTask, prepared *preparedWorkspace) {
	if prepared == nil {
		return
	}
	task.workspace = prepared.workspace
	task.events = prepared.events
}

func forwardActivatedSession(a *Actor, task *runningTask) {
	go a.forwardSession(task.taskID, task.handle.Events(), task.handle.Result())
}
