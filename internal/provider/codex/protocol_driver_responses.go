package codex

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
		// Send thread/start (or thread/resume).
		method := "thread/start"
		params := map[string]any{}
		if d.req.ResumeSessionID != "" {
			method = "thread/resume"
			params["threadId"] = d.req.ResumeSessionID
		} else if d.req.Model != "" {
			params["model"] = d.req.Model
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
