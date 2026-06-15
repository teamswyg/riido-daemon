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
	d, _ := NewProtocolDriver(agentbridge.StartRequest{Model: "gpt-5.5"})
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
	if !strings.Contains(string(second), `"model":"gpt-5.5"`) {
		t.Fatalf("thread/start missing model: %q", second)
	}
}

// --- B3: thread response → turn/start ---

func TestCodexProtocolDriverThreadResponseWritesTurnStart(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{Prompt: "do the thing", Model: "gpt-5.5"})
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
	if !strings.Contains(string(frame), `"model":"gpt-5.5"`) {
		t.Fatalf("turn/start missing model: %q", frame)
	}
}
