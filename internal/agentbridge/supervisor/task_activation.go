package supervisor

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) prepareAndSubmit(ctx context.Context, status runtimeactor.Status, rt *runtimeactor.Actor, req *bridge.TaskRequest) {
	prepared, err := a.prepareWorkspace(ctx, status, req)
	if err != nil {
		a.forwardActivation(taskActivationMsg{taskID: req.ID, err: err})
		return
	}
	handle, err := rt.Submit(ctx, *req)
	if err != nil {
		a.forwardActivation(taskActivationMsg{taskID: req.ID, prepared: prepared, err: err})
		return
	}
	a.forwardActivation(taskActivationMsg{taskID: req.ID, prepared: prepared, handle: handle})
}

func (a *Actor) forwardActivation(msg taskActivationMsg) {
	select {
	case a.mailbox <- envelope{taskActivation: &msg}:
	case <-a.stoppedCh:
	}
}

func resultForActivationError(err error) agentbridge.Result {
	status := agentbridge.ResultFailed
	if errors.Is(err, context.Canceled) || errors.Is(err, ErrStopped) {
		status = agentbridge.ResultCancelled
	}
	if errors.Is(err, errAssignmentWorktreeBlocked) {
		status = agentbridge.ResultBlocked
	}
	return agentbridge.Result{Status: status, Error: err.Error()}
}
