package codex

import (
	"context"
	"encoding/json"
	"fmt"

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
	initialized bool
	threadID    string
	turnStarted bool
}

type pendingRequest struct {
	method string
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
	_, err := d.sendRequest(ctx, io, "initialize", map[string]any{
		"clientInfo": map[string]any{"name": "riido", "version": "0.0.0"},
	})
	return err
}

// OnRaw routes a parser RawEvent through either the handshake state
// machine (for "response") or the existing Translate (for notifications
// and server_requests).
func (d *protocolDriver) OnRaw(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "response" {
		return d.handleResponse(ctx, raw, io)
	}
	if raw.Type == "error" {
		// Mirror the existing Translate behavior for error responses.
		id, _ := rpcID(raw.Payload)
		delete(d.pending, id)
		return Translate(raw)
	}
	// Notifications + server_requests + malformed + stderr: fall through
	// to existing Translate. This keeps the Codex translator the single
	// source of provider→IR mappings.
	return Translate(raw)
}

// OnProcessExit emits one Error event per still-pending request so
// callers never block forever waiting for a response that will never
// arrive. Also clears the pending map (cleanup).
func (d *protocolDriver) OnProcessExit(_ context.Context, status agentbridge.ProcessExitStatus, _ agentbridge.ProtocolIO) ([]agentbridge.Event, error) {
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

// --- handshake state machine (NOT a RunState FSM — just transport progress) ---

func (d *protocolDriver) handleResponse(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	id, ok := rpcID(raw.Payload)
	if !ok {
		return nil, nil, nil
	}
	pr, known := d.pending[id]
	if !known {
		// Stray response — no-op.
		return nil, nil, nil
	}
	delete(d.pending, id)

	switch pr.method {
	case "initialize":
		// Send "initialized" notification (no id, no response expected).
		if err := d.sendNotification(ctx, io, "initialized", map[string]any{}); err != nil {
			return nil, nil, err
		}
		d.initialized = true
		// Send thread/start (or thread/resume).
		method := "thread/start"
		params := map[string]any{}
		if d.req.ResumeSessionID != "" {
			method = "thread/resume"
			params["threadId"] = d.req.ResumeSessionID
		}
		if _, err := d.sendRequest(ctx, io, method, params); err != nil {
			return nil, nil, err
		}

	case "thread/start", "thread/resume":
		result := mapField(raw.Payload, "result")
		d.threadID = threadIDFromResult(result)
		// Surface as an agentbridge SessionIdentified so downstream
		// consumers see the session id without waiting for the
		// "thread_started" notification (which Codex may or may not
		// also emit).
		events := []agentbridge.Event{}
		if d.threadID != "" {
			events = append(events, agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: d.threadID})
		}
		// Kick off turn/start.
		params := map[string]any{}
		if d.threadID != "" {
			params["threadId"] = d.threadID
		}
		if d.req.Prompt != "" {
			params["input"] = []map[string]any{{"type": "text", "text": d.req.Prompt}}
		}
		if _, err := d.sendRequest(ctx, io, "turn/start", params); err != nil {
			return nil, nil, err
		}
		return events, nil, nil

	case "turn/start":
		d.turnStarted = true
		// The Codex server may emit turn_started notification later;
		// translate handles that. Here we emit a Lifecycle so the
		// reducer reflects "running" even if we never see the
		// notification.
		return []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}}, nil, nil
	}
	return nil, nil, nil
}

// --- transport helpers ---

func (d *protocolDriver) sendRequest(ctx context.Context, io agentbridge.ProtocolIO, method string, params map[string]any) (int64, error) {
	d.nextID++
	id := d.nextID
	d.pending[id] = pendingRequest{method: method}
	frame := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		frame["params"] = params
	}
	return id, d.writeFrame(ctx, io, frame)
}

func (d *protocolDriver) sendNotification(ctx context.Context, io agentbridge.ProtocolIO, method string, params map[string]any) error {
	frame := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		frame["params"] = params
	}
	return d.writeFrame(ctx, io, frame)
}

func (d *protocolDriver) writeFrame(ctx context.Context, io agentbridge.ProtocolIO, frame map[string]any) error {
	b, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("codex driver: marshal %v: %w", frame["method"], err)
	}
	b = append(b, '\n')
	return io.WriteStdin(ctx, b)
}

// rpcID extracts the JSON-RPC "id" field as int64. JSON numbers come
// back as float64; we accept int / int64 too for callers that build
// payloads directly without a JSON roundtrip (tests, internal helpers).
func rpcID(p map[string]any) (int64, bool) {
	switch v := p["id"].(type) {
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int64:
		return v, true
	}
	return 0, false
}

// mapField returns p[key] as a map, or nil.
func mapField(p map[string]any, key string) map[string]any {
	if p == nil {
		return nil
	}
	m, _ := p[key].(map[string]any)
	return m
}

func threadIDFromResult(result map[string]any) string {
	if id := stringField(result, "thread_id"); id != "" {
		return id
	}
	thread := mapField(result, "thread")
	if id := stringField(thread, "id"); id != "" {
		return id
	}
	return stringField(thread, "sessionId")
}
