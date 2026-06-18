package session

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type sessionRunner struct {
	ctx  context.Context
	cfg  Config
	sess *Session
	proc process.RunningProcess

	parser    agentbridge.Parser
	state     agentbridge.State
	telemetry *agentbridge.TelemetryParser
	io        *protocolIOImpl

	startedAt time.Time
	stdoutCh  <-chan []byte
	stderrCh  <-chan []byte
	exitedCh  <-chan process.ExitStatus

	hardTimer *time.Timer
	hardC     <-chan time.Time
	idleTimer *time.Timer
	idleC     <-chan time.Time

	deferredExit *process.ExitStatus
}
