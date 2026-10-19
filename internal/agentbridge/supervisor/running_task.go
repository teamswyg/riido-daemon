package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type runningTask struct {
	recovery      *assignment.RecoveryContext
	recoveryScope string

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

	workspace           *workdir.Workspace
	execution           *execution.Prepared
	request             *bridge.TaskRequest
	canResume           bool
	continuations       int
	verificationPending bool
	executionVerified   bool
	priorUsage          agentbridge.Usage
	initialStartedAt    time.Time
	events              *workspaceEventContext

	nativeSessionID           string
	sessionBindingErr         error
	terminalResult            *agentbridge.Result
	terminalDeliveryAttempted bool
	pendingEvents             []agentbridge.Event
	supersededEvents          []agentbridge.Event
}

type preparedWorkspace struct {
	execution *execution.Prepared
	request   *bridge.TaskRequest
	canResume bool
	workspace *workdir.Workspace
	events    *workspaceEventContext
}
