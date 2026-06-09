package codex

import (
	"context"
	"encoding/json"
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
	req      agentbridge.StartRequest
	turnOnly bool

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

// NewTurnProtocolDriver returns a driver for an already-initialized Codex
// app-server process. It starts or resumes a provider thread and then starts a
// turn, but it does not send initialize/initialized.
func NewTurnProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return NewTurnProtocolDriverWithIDSeed(req, 1)
}

// NewTurnProtocolDriverWithIDSeed is the persistent-process variant. The
// caller supplies a monotonically increasing seed so late JSON-RPC responses
// from an earlier turn cannot collide with the next turn's pending ids.
func NewTurnProtocolDriverWithIDSeed(req agentbridge.StartRequest, seed int64) (agentbridge.ProtocolDriver, error) {
	if seed <= 0 {
		seed = 1
	}
	return &protocolDriver{
		req:      req,
		turnOnly: true,
		nextID:   seed - 1,
		pending:  map[int64]pendingRequest{},
	}, nil
}

// OnStart writes the initialize request.
func (d *protocolDriver) OnStart(ctx context.Context, io agentbridge.ProtocolIO) error {
	if d.turnOnly {
		return d.startThread(ctx, io)
	}
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
		id, _ := rpcID(raw.Payload)
		if pr, known := d.pending[id]; known {
			delete(d.pending, id)
			return d.failedEvents("codex " + pr.method + " rpc error: " + codexRPCErrorMessage(raw.Payload)), nil, nil
		}
		d.recordRuntimeError(codexRPCErrorMessage(raw.Payload))
		events, cmds, err := Translate(raw)
		d.observeEvents(events)
		return events, cmds, err
	}
	if strings.HasPrefix(raw.Type, "notification:") {
		method := strings.TrimPrefix(raw.Type, "notification:")
		if method == "error" {
			errText := codexNotificationErrorMessage(params(raw))
			d.recordRuntimeError(errText)
			return []agentbridge.Event{{Kind: agentbridge.EventError, Err: errText}}, nil, nil
		}
		if method == "turn_started" || method == "turn/started" {
			d.turnStarted = true
		}
		if method == "turn_completed" || method == "turn/completed" {
			p := params(raw)
			if d.lastRuntimeError != "" && !d.sawAssistantOutput && stringField(p, "output") == "" {
				return d.failedEvents(d.lastRuntimeError), nil, nil
			}
		}
		// Newer codex app-server builds signal turn end via thread/status/changed
		// (the thread returns to a terminal/idle status) instead of
		// turn/completed. Without this the run never receives a completion and
		// fails with "codex unknown notification: thread/status/changed".
		if method == "thread/status/changed" || method == "thread_status_changed" {
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
		}
	}
}

func (d *protocolDriver) recordRuntimeError(message string) {
	if message == "" {
		message = "codex runtime error"
	}
	d.lastRuntimeError = message
}

func (d *protocolDriver) failedEvents(message string) []agentbridge.Event {
	if message == "" {
		message = "codex runtime error"
	}
	return []agentbridge.Event{
		{Kind: agentbridge.EventError, Err: message},
		{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultFailed, Error: message}},
	}
}

// threadStatusEvents maps a codex app-server thread/status/changed notification
// to run-scope events. The new codex protocol uses thread status transitions
// (the thread returns to idle/completed when a turn finishes) instead of a
// turn/completed notification. Completion is gated on turnStarted so an initial
// idle status cannot end the run before any work happens; unknown statuses are
// logged (never fail the run).
func (d *protocolDriver) threadStatusEvents(p map[string]any) []agentbridge.Event {
	status := strings.ToLower(strings.TrimSpace(codexThreadStatus(p)))
	switch {
	case codexStatusIsError(status):
		return d.failedEvents("codex thread status: " + status)
	case codexStatusIsTerminal(status):
		if !d.turnStarted {
			return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex thread status: " + status + " (no active turn)"}}
		}
		d.turnStarted = false
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}
	case codexStatusIsActive(status):
		d.turnStarted = true
		return []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}}
	default:
		return []agentbridge.Event{{Kind: agentbridge.EventLog, Text: "codex thread status changed: " + status}}
	}
}

// codexThreadStatus extracts the status string from a thread/status/changed
// payload, tolerating a few shapes (flat string, nested object, or "state").
func codexThreadStatus(p map[string]any) string {
	if s := stringField(p, "status"); s != "" {
		return s
	}
	if s := stringField(mapField(p, "status"), "type"); s != "" {
		return s
	}
	if s := stringField(mapField(p, "thread"), "status"); s != "" {
		return s
	}
	return stringField(p, "state")
}

func codexStatusIsTerminal(status string) bool {
	switch status {
	case "idle", "completed", "complete", "finished", "done", "ready", "succeeded":
		return true
	}
	return false
}

func codexStatusIsError(status string) bool {
	switch status {
	case "error", "errored", "failed", "aborted", "cancelled", "canceled", "interrupted":
		return true
	}
	return false
}

func codexStatusIsActive(status string) bool {
	switch status {
	case "running", "active", "in_progress", "working", "streaming", "thinking", "busy", "generating", "turn_running":
		return true
	}
	return false
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
		if err := d.startThread(ctx, io); err != nil {
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
		if d.req.Model != "" {
			params["model"] = d.req.Model
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

func (d *protocolDriver) startThread(ctx context.Context, io agentbridge.ProtocolIO) error {
	method := "thread/start"
	params := map[string]any{}
	if d.req.ResumeSessionID != "" {
		method = "thread/resume"
		params["threadId"] = d.req.ResumeSessionID
	} else if d.req.Model != "" {
		params["model"] = d.req.Model
	}
	_, err := d.sendRequest(ctx, io, method, params)
	return err
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

func codexRPCErrorMessage(payload map[string]any) string {
	if msg := errMessage(payload); msg != "" {
		return msg
	}
	return "codex rpc error"
}

func codexNotificationErrorMessage(p map[string]any) string {
	if msg := stringField(p, "message"); msg != "" {
		return msg
	}
	if msg := stringField(p, "detail"); msg != "" {
		return msg
	}
	if errText := stringField(p, "error"); errText != "" {
		return errText
	}
	errMap := mapField(p, "error")
	if msg := stringField(errMap, "message"); msg != "" {
		return msg
	}
	return "codex runtime error"
}
