package supervisor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

type envelope struct {
	taskActivation *taskActivationMsg
	taskEvent      *taskEventMsg
	taskResult     *taskResultMsg
	cancel         *cancelMsg
}

type taskActivationMsg struct {
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
