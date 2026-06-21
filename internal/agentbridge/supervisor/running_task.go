package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type runningTask struct {
	taskID                string
	provider              string
	runtimeID             string
	capabilityFingerprint string
	ctx                   context.Context
	report                controlplane.TaskReportContext
	runtime               *runtimeactor.Actor
	handle                *runtimeactor.SessionHandle
	cancel                context.CancelFunc
	cancelCause           error

	workspace *workdir.Workspace
	events    *workspaceEventContext

	terminalResult *agentbridge.Result
}

type preparedWorkspace struct {
	workspace *workdir.Workspace
	events    *workspaceEventContext
}
