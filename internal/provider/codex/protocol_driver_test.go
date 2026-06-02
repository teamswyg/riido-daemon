package codex

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// recordingIO is an agentbridge.ProtocolIO for driver tests. Writes land in
// a channel so we can assert ordering and content without locks.
type recordingIO struct {
	writes chan []byte
	closed bool
}

func newRecordingIO() *recordingIO { return &recordingIO{writes: make(chan []byte, 16)} }

func (r *recordingIO) WriteStdin(_ context.Context, b []byte) error {
	r.writes <- append([]byte(nil), b...)
	return nil
}

func (r *recordingIO) CloseStdin(_ context.Context) error {
	r.closed = true
	return nil
}

func (r *recordingIO) next(t *testing.T, deadline time.Duration) []byte {
	t.Helper()
	select {
	case b := <-r.writes:
		return b
	case <-time.After(deadline):
		t.Fatalf("expected stdin write within %v", deadline)
		return nil
	}
}

// makeResponse builds a JSON-RPC response RawEvent the driver will see.
func makeResponse(id int64, result map[string]any) agentbridge.RawEvent {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	b, _ := json.Marshal(payload)
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "response",
		Payload: payload,
		Bytes:   b,
	}
}

// makeNotification builds a JSON-RPC notification RawEvent.
func makeNotification(method string, params map[string]any) agentbridge.RawEvent {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	b, _ := json.Marshal(payload)
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "notification:" + method,
		Payload: payload,
		Bytes:   b,
	}
}

func makeServerRequest(id int64, method string, params map[string]any) agentbridge.RawEvent {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	b, _ := json.Marshal(payload)
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "server_request:" + method,
		Payload: payload,
		Bytes:   b,
	}
}

// --- B1: OnStart writes initialize ---

func TestCodexProtocolDriverWritesInitializeOnStart(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	if err := d.OnStart(context.Background(), io); err != nil {
		t.Fatalf("OnStart: %v", err)
	}
	b := io.next(t, time.Second)
	if !strings.Contains(string(b), `"method":"initialize"`) {
		t.Fatalf("first frame is not initialize: %q", b)
	}
	if !strings.Contains(string(b), `"clientInfo"`) {
		t.Fatalf("initialize missing clientInfo: %q", b)
	}
}

// --- B2: initialize response → initialized + thread/start ---

func TestCodexProtocolDriverInitializeResponseWritesInitializedAndThreadStart(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	if err := d.OnStart(context.Background(), io); err != nil {
		t.Fatal(err)
	}
	_ = io.next(t, time.Second) // initialize frame

	// Feed initialize response (id=1).
	_, _, err := d.OnRaw(context.Background(), makeResponse(1, map[string]any{"server": "codex"}), io)
	if err != nil {
		t.Fatalf("OnRaw: %v", err)
	}

	// Expect "initialized" notification next.
	first := io.next(t, time.Second)
	if !strings.Contains(string(first), `"method":"initialized"`) {
		t.Fatalf("expected initialized notification, got %q", first)
	}
	// Then thread/start request.
	second := io.next(t, time.Second)
	if !strings.Contains(string(second), `"method":"thread/start"`) {
		t.Fatalf("expected thread/start, got %q", second)
	}
}

// --- B3: thread response → turn/start ---

func TestCodexProtocolDriverThreadResponseWritesTurnStart(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{Prompt: "do the thing"})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second) // initialize

	// Initialize response → initialized + thread/start.
	_, _, _ = d.OnRaw(context.Background(), makeResponse(1, nil), io)
	_ = io.next(t, time.Second) // initialized
	_ = io.next(t, time.Second) // thread/start

	// thread/start response with current Codex app-server nested thread id.
	_, _, _ = d.OnRaw(context.Background(), makeResponse(2, map[string]any{"thread": map[string]any{"id": "th-xyz"}}), io)

	// Next frame must be turn/start carrying the prompt input and threadId.
	frame := io.next(t, time.Second)
	if !strings.Contains(string(frame), `"method":"turn/start"`) {
		t.Fatalf("expected turn/start, got %q", frame)
	}
	if !strings.Contains(string(frame), "th-xyz") {
		t.Fatalf("turn/start missing thread id: %q", frame)
	}
	if !strings.Contains(string(frame), "do the thing") {
		t.Fatalf("turn/start missing input text: %q", frame)
	}
	if !strings.Contains(string(frame), `"input"`) {
		t.Fatalf("turn/start missing input array: %q", frame)
	}
}

// --- B4: streaming agent_message → TextDelta ---

func TestCodexProtocolDriverStreamingNotificationTranslatesToEvents(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	events, _, err := d.OnRaw(context.Background(), makeNotification("agent_message", map[string]any{"text": "hi"}), io)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta || events[0].Text != "hi" {
		t.Fatalf("text delta: %+v", events)
	}
}

// --- B5: approve_command → ToolApprovalNeeded ---

func TestCodexProtocolDriverApprovalRequestTranslatesToApprovalEvent(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	events, _, err := d.OnRaw(context.Background(),
		makeServerRequest(42, "approve_command", map[string]any{"id": "cmd-1", "command": "ls"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("approval: %+v", events)
	}
	if events[0].Tool.ID != "cmd-1" {
		t.Fatalf("tool ref: %+v", events[0].Tool)
	}
}

// --- B6: turn_completed → ResultCompleted ---

func TestCodexProtocolDriverTurnCompletedReturnsResult(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	events, _, err := d.OnRaw(context.Background(),
		makeNotification("turn_completed", map[string]any{"output": "all done"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("no events")
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}
}

func TestCodexProtocolDriverRPCErrorFailsPendingRequest(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second) // initialize pending id=1

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      int64(1),
		"error":   map[string]any{"message": "missing auth"},
	}
	events, _, err := d.OnRaw(context.Background(), agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "error",
		Payload: payload,
	}, io)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("expected failed result for rpc error, got %+v", events)
	}
	if !strings.Contains(last.Result.Error, "initialize") || !strings.Contains(last.Result.Error, "missing auth") {
		t.Fatalf("missing rpc error context: %+v", last.Result)
	}
}

func TestCodexProtocolDriverErrorNotificationThenEmptyCompletionFails(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	events, _, err := d.OnRaw(context.Background(),
		makeNotification("error", map[string]any{"message": "failed to connect to websocket: 401 Unauthorized"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventError {
		t.Fatalf("expected non-terminal error event first, got %+v", events)
	}

	events, _, err = d.OnRaw(context.Background(),
		makeNotification("turn_completed", map[string]any{"output": ""}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("expected failed result after empty completion, got %+v", events)
	}
	if !strings.Contains(last.Result.Error, "401 Unauthorized") {
		t.Fatalf("expected retained runtime error, got %+v", last.Result)
	}
}

func TestCodexProtocolDriverErrorNotificationCanRecoverWithText(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	_, _, _ = d.OnRaw(context.Background(),
		makeNotification("error", map[string]any{"message": "temporary stream disconnect"}),
		io,
	)
	events, _, err := d.OnRaw(context.Background(),
		makeNotification("agent_message", map[string]any{"text": "작업 완료"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta {
		t.Fatalf("expected text delta, got %+v", events)
	}

	events, _, err = d.OnRaw(context.Background(),
		makeNotification("turn_completed", map[string]any{"output": ""}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("expected recovered completion, got %+v", events)
	}
}

func TestCodexProtocolDriverProcessExitAfterOnlyRuntimeErrorFails(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	_, _, _ = d.OnRaw(context.Background(),
		makeNotification("error", map[string]any{"error": map[string]any{"message": "provider auth failed"}}),
		io,
	)
	events, err := d.OnProcessExit(context.Background(), agentbridge.ProcessExitStatus{Code: 0}, io)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("expected failed process-exit result, got %+v", events)
	}
	if !strings.Contains(last.Result.Error, "provider auth failed") {
		t.Fatalf("expected retained error message, got %+v", last.Result)
	}
}

// --- B7: process exit fails pending requests ---

func TestCodexProtocolDriverProcessExitFailsPendingRequests(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second) // initialize sent → pending request id=1 in flight

	// Send thread/start before initialize response arrives so that
	// two pending requests are live.
	_, _, _ = d.OnRaw(context.Background(), makeResponse(1, nil), io)
	_ = io.next(t, time.Second) // initialized
	_ = io.next(t, time.Second) // thread/start → pending id=2

	exitEvents, err := d.OnProcessExit(context.Background(),
		agentbridge.ProcessExitStatus{Code: 137}, io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(exitEvents) == 0 {
		t.Fatal("expected at least one event for pending request cleanup")
	}
	// At least one Error event should reference "pending" semantics.
	gotError := false
	for _, ev := range exitEvents {
		if ev.Kind == agentbridge.EventError {
			gotError = true
		}
	}
	if !gotError {
		t.Fatalf("expected an Error event after process exit, got %+v", exitEvents)
	}
}

// --- Integration: agentbridge.ProtocolDriver contract is satisfied. ---

func TestCodexProtocolDriverImplementsSessionInterface(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	if d == nil {
		t.Fatal("driver is nil")
	}
	// Compile-time check: assignable to agentbridge.ProtocolDriver.
	var _ agentbridge.ProtocolDriver = d
}
