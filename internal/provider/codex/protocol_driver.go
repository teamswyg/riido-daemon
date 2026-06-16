package codex

import (
	"context"
	"fmt"
	"strings"

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

// OnRaw routes a parser RawEvent through either the handshake state
// machine (for "response") or the existing Translate (for notifications
// and server_requests).
func (d *protocolDriver) OnRaw(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if rawFrameType(raw.Type) == rawFrameResponse {
		return d.handleResponse(ctx, raw, io)
	}
	if rawFrameType(raw.Type) == rawFrameError {
		id, _ := rpcID(raw.Payload)
		if pr, known := d.pending[id]; known {
			delete(d.pending, id)
			return d.failedEvents("codex " + string(pr.method) + " rpc error: " + codexRPCErrorMessage(raw.Payload)), nil, nil
		}
		d.recordRuntimeError(codexRPCErrorMessage(raw.Payload))
		events, cmds, err := Translate(raw)
		d.observeEvents(events)
		return events, cmds, err
	}
	if after, ok := strings.CutPrefix(raw.Type, rawFrameNotificationPrefix); ok {
		method := codexMethod(after)
		if method == codexMethodError {
			errText := codexNotificationErrorMessage(params(raw))
			d.recordRuntimeError(errText)
			return []agentbridge.Event{{Kind: agentbridge.EventError, Err: errText}}, nil, nil
		}
		if method == codexMethodTurnStarted || method == codexMethodTurnStartedSlash {
			d.turnStarted = true
		}
		if method == codexMethodTurnCompleted || method == codexMethodTurnCompleteSlash {
			p := params(raw)
			if d.lastRuntimeError != "" && !d.sawAssistantOutput && stringField(p, "output") == "" {
				return d.failedEvents(d.lastRuntimeError), nil, nil
			}
		}
		// Newer codex app-server builds signal turn end via thread/status/changed
		// (the thread returns to a terminal/idle status) instead of
		// turn/completed. Without this the run never receives a completion and
		// fails with "codex unknown notification: thread/status/changed".
		if method == codexMethodThreadStatusChanged || method == codexMethodThreadStatusAlt {
			events := d.threadStatusEvents(params(raw))
			d.observeEvents(events)
			return events, nil, nil
		}
	}
	// Notifications + server_requests + malformed + stderr: fall through
	// to existing Translate. This keeps the Codex translator the single
	// source of provider→IR mappings.
	events, cmds, err := Translate(raw)
	d.observeEvents(events)
	return events, cmds, err
}

// OnProcessExit emits one Error event per still-pending request so
// callers never block forever waiting for a response that will never
// arrive. Also clears the pending map (cleanup).
func (d *protocolDriver) OnProcessExit(_ context.Context, status agentbridge.ProcessExitStatus, _ agentbridge.ProtocolIO) ([]agentbridge.Event, error) {
	if d.lastRuntimeError != "" && !d.sawAssistantOutput {
		return d.failedEvents(d.lastRuntimeError), nil
	}
	if len(d.pending) == 0 {
		return nil, nil
	}
	out := make([]agentbridge.Event, 0, len(d.pending))
	for id, pr := range d.pending {
		out = append(out, agentbridge.Event{
			Kind: agentbridge.EventError,
			Err:  fmt.Sprintf("codex: pending RPC request id=%d method=%s cancelled by process exit code=%d", id, pr.method, status.Code),
		})
	}
	d.pending = map[int64]pendingRequest{}
	return out, nil
}

// OnClose releases any final state. The pending map is already cleared
// by OnProcessExit; if the session terminates for a non-exit reason
// (Cancel/Timeout) we still want pending entries dropped so future
// re-use can't leak them.
func (d *protocolDriver) OnClose(_ context.Context, _ agentbridge.ProtocolIO) error {
	d.pending = nil
	return nil
}

func (d *protocolDriver) observeEvents(events []agentbridge.Event) {
	for _, ev := range events {
		switch ev.Kind {
		case agentbridge.EventTextDelta:
			if strings.TrimSpace(ev.Text) != "" {
				d.sawAssistantOutput = true
			}
		case agentbridge.EventResult:
			if strings.TrimSpace(ev.Result.Output) != "" {
				d.sawAssistantOutput = true
			}
		default:
		}
	}
}

func (d *protocolDriver) recordRuntimeError(message string) {
	if message == "" {
		message = "codex runtime error"
	}
	d.lastRuntimeError = message
}
