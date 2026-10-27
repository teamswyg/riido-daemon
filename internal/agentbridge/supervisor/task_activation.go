package supervisor

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) prepareAndSubmit(ctx context.Context, status runtimeactor.Status, rt *runtimeactor.Actor, req *bridge.TaskRequest, work *executionproof.HostWork) {
	a.recoverAndSubmit(ctx, status, rt, req, work)
}

func (a *Actor) forwardActivation(msg taskActivationMsg) {
	// Worker completion must precede mailbox visibility and possible terminal sealing.
	msg.hostWork.Finished()
	select {
	case a.mailbox <- envelope{taskActivation: &msg}:
	case <-a.stoppedCh:
	}
}

func resultForActivationError(err error) agentbridge.Result {
	status := agentbridge.ResultFailed
	var readyErr *readyAcknowledgementFailure
	if errors.As(err, &readyErr) && readyErr.blocked {
		status = agentbridge.ResultBlocked
	}
	var recoveryErr *recoveryFailure
	if errors.As(err, &recoveryErr) && recoveryErr.blocked {
		status = agentbridge.ResultBlocked
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, ErrStopped) {
		status = agentbridge.ResultCancelled
	}
	if errors.Is(err, errAssignmentWorktreeBlocked) || errors.Is(err, errNativeSessionBinding) {
		status = agentbridge.ResultBlocked
	}
	code, action := "provider_failed", "check_provider"
	if readyErr != nil {
		code, action = "unknown", "review_result"
	}
	if errors.Is(err, errWorkspacePreparation) {
		code, action = "workspace_preparation_failed", "check_workspace"
	}
	if errors.Is(err, errNativeSessionBinding) {
		code, action = "session_scope_conflict", "review_session"
	}
	return agentbridge.Result{Status: status, Error: err.Error(), ExecutionContext: &assignment.ExecutionContext{WorkspaceKind: "unknown", VerificationStatus: "unknown", DiagnosticCode: code, NextAction: action}}
}
