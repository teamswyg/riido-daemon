package codex

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// protocolDriver implements agentbridge.ProtocolDriver for Codex's
// app-server JSON-RPC transport.
//
// The driver autonomously walks the handshake the daemon needs to send
// to bring a Codex session live:
//
//	OnStart                                  → initialize
//	OnRaw(response to initialize)            → initialized notification
//	                                         + thread/start
//	OnRaw(response to thread/start)          → turn/start
//	OnRaw(response to turn/start)            → (just consume; turn is live)
//	OnRaw(notifications)                     → existing Translate routing
//	OnRaw(server_request approve_command)    → existing Translate routing
//	                                           (returns ToolApprovalNeeded)
//	OnProcessExit                            → emit Error event for each
//	                                           still-pending request
//
// All internal bookkeeping (next request id, pending map, handshake
// progress booleans, captured thread id) lives on the driver struct.
// The session actor is the SOLE caller, so no mutex.
//
// NOTE: this is NOT a per-provider RunState FSM. It is transport
// progress tracking that the agentbridge core neither sees nor cares
// about. The canonical RunState lives in agentbridge.State, owned by
// the reducer that consumes the events returned by OnRaw.
type protocolDriver struct {
	req agentbridge.StartRequest

	nextID  int64
	pending map[int64]pendingRequest

	// Handshake progress flags. Single-goroutine access (session actor).
	initialized        bool
	threadID           string
	turnStarted        bool
	sawAssistantOutput bool
	lastRuntimeError   string
}

type pendingRequest struct {
	method codexMethod
}

// NewProtocolDriver returns a provider-neutral ProtocolDriver that drives the
// Codex app-server JSON-RPC handshake.
func NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return &protocolDriver{
		req:     req,
		pending: map[int64]pendingRequest{},
	}, nil
}

// OnStart writes the initialize request.
func (d *protocolDriver) OnStart(ctx context.Context, io agentbridge.ProtocolIO) error {
	_, err := d.sendRequest(ctx, io, codexMethodInitialize, map[string]any{
		"clientInfo": map[string]any{"name": "riido", "version": "0.0.0"},
	})
	return err
}
