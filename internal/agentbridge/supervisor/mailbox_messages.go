package supervisor

import (
	"context"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

type envelope struct {
	taskActivation   *taskActivationMsg
	taskEvent        *taskEventMsg
	taskReport       *taskEventMsg
	taskResult       *taskResultMsg
	taskVerification *taskVerificationMsg
	cancel           *cancelMsg
}

type taskActivationMsg struct {
	claimContext  context.Context
	runtime       *runtimeactor.Actor
	status        runtimeactor.Status
	recovery      *assignment.RecoveryContext
	recoveryScope string

	taskID   string
	prepared *preparedWorkspace
	handle   *runtimeactor.SessionHandle
	err      error
}

type taskEventMsg struct {
	taskID string
	event  agentbridge.Event
}

type taskResultMsg struct {
	taskID string
	result agentbridge.Result
}

type cancelMsg struct {
	taskID string
	cause  error
}
