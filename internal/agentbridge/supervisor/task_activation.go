package supervisor

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) prepareAndSubmit(ctx context.Context, status runtimeactor.Status, rt *runtimeactor.Actor, req *bridge.TaskRequest) {
	prepared, err := a.prepareWorkspace(ctx, status, req)
	if err != nil {
		a.forwardActivation(taskActivationMsg{taskID: req.ID, err: fmt.Errorf("%w: %w", errWorkspacePreparation, err)})
		return
	}
	if prepared != nil {
		snapshot := *req
		prepared.request = &snapshot
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
	if errors.Is(err, errAssignmentWorktreeBlocked) || errors.Is(err, errNativeSessionBinding) {
		status = agentbridge.ResultBlocked
	}
	code, action := "provider_failed", "check_provider"
	if errors.Is(err, errWorkspacePreparation) {
		code, action = "workspace_preparation_failed", "check_workspace"
	}
	if errors.Is(err, errNativeSessionBinding) {
		code, action = "session_scope_conflict", "review_session"
	}
	return agentbridge.Result{Status: status, Error: err.Error(), ExecutionContext: &assignment.ExecutionContext{WorkspaceKind: "unknown", VerificationStatus: "unknown", DiagnosticCode: code, NextAction: action}}
}
