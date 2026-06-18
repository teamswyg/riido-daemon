package runtimeactor

import (
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// Actor is the runtime tier actor.
type Actor struct {
	cfg Config

	// Single owning goroutine writes to capabilities/inFlight/...; all
	// public methods send to mailbox channels.
	mailbox  chan envelope
	statusCh chan statusMsg
	// stopReqCh carries the requested shutdown authority level. Stop callers
	// do a non-blocking send; forced requests can escalate during drain.
	stopReqCh chan lifecycle.ShutdownLevel
	stoppedCh chan struct{}
	stopErrCh chan error
	startedCh chan struct{}
	startedAt time.Time
}

// runningTask is the actor's per-task bookkeeping.
type runningTask struct {
	taskID   string
	provider string
	handle   *SessionHandle
}
