package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// envelope is the actor's mailbox shape: one message kind per call.
type envelope struct {
	submit *submitMsg
	cancel *cancelMsg
}

type submitMsg struct {
	ctx   context.Context
	req   bridge.TaskRequest
	reply chan submitReply
}

type submitReply struct {
	handle *SessionHandle
	err    error
}

type cancelMsg struct {
	ctx    context.Context
	taskID string
	reason string
	reply  chan error
}

type statusMsg struct {
	ctx   context.Context
	reply chan statusReply
}

type statusReply struct {
	status Status
	hb     Heartbeat
}
