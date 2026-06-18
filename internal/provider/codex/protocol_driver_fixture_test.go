package codex

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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

func makeResponse(id int64, result map[string]any) agentbridge.RawEvent {
	payload := map[string]any{"jsonrpc": "2.0", "id": id, "result": result}
	b, _ := json.Marshal(payload)
	return agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: "response", Payload: payload, Bytes: b}
}

func makeNotification(method string, params map[string]any) agentbridge.RawEvent {
	payload := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	b, _ := json.Marshal(payload)
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "notification:" + method,
		Payload: payload,
		Bytes:   b,
	}
}

func makeServerRequest(id int64, method string, params map[string]any) agentbridge.RawEvent {
	payload := map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}
	b, _ := json.Marshal(payload)
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "server_request:" + method,
		Payload: payload,
		Bytes:   b,
	}
}
