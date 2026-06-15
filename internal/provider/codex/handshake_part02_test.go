package codex

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// --- helpers ---

// jsonline serializes one JSON-RPC frame (one line, '\n' terminated).
func jsonline(t *testing.T, m map[string]any) []byte {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return append(b, '\n')
}

// mustWriteJSONRPC simulates the daemon writing a JSON-RPC frame to the
// codex process's stdin. We do not assert delivery — the FakeRunning's
// stdin channel records every write for later inspection.
func mustWriteJSONRPC(t *testing.T, r *process.FakeRunning, m map[string]any) {
	t.Helper()
	if err := r.WriteStdin(jsonline(t, m)); err != nil {
		t.Fatalf("WriteStdin: %v", err)
	}
}

// drainHandshake pulls available stdout chunks, feeds the parser, runs
// Translate, optionally routes responses to the RPC actor, and invokes
// the predicate on each step. Returns once the predicate is satisfied
// or the deadline expires.
func drainHandshake(
	t *testing.T,
	parser agentbridge.Parser,
	stdout <-chan []byte,
	rpc *RPCActor,
	deadline time.Duration,
	predicate func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool,
) {
	t.Helper()
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		select {
		case chunk, ok := <-stdout:
			if !ok {
				return
			}
			raws, err := parser.FeedStdout(chunk)
			if err != nil {
				t.Fatalf("parser: %v", err)
			}
			for _, raw := range raws {
				if raw.Type == "response" {
					if id, ok := rpcID(raw.Payload); ok {
						rpc.Resolve(id, mapField(raw.Payload, "result"), nil)
					}
				}
				if raw.Type == "error" {
					if id, ok := rpcID(raw.Payload); ok {
						rpc.Resolve(id, nil, jsonRPCError(raw.Payload))
					}
				}
				evs, _, err := Translate(raw)
				if err != nil {
					t.Fatalf("translate: %v", err)
				}
				if predicate(raw, evs) {
					return
				}
			}
		case <-time.After(50 * time.Millisecond):
			// poll
		}
	}
}

// routeRPC is a thin wrapper around drainHandshake when the caller only
// cares about routing one specific response id.
func routeRPC(t *testing.T, parser agentbridge.Parser, stdout <-chan []byte, rpc *RPCActor, deadline time.Duration, awaitID int64) {
	t.Helper()
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		select {
		case chunk, ok := <-stdout:
			if !ok {
				return
			}
			raws, _ := parser.FeedStdout(chunk)
			for _, raw := range raws {
				if raw.Type != "response" {
					continue
				}
				id, hasID := rpcID(raw.Payload)
				if !hasID {
					continue
				}
				rpc.Resolve(id, mapField(raw.Payload, "result"), nil)
				if id == awaitID {
					return
				}
			}
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func jsonRPCError(p map[string]any) error {
	e, _ := p["error"].(map[string]any)
	msg, _ := e["message"].(string)
	if msg == "" {
		msg = "unknown rpc error"
	}
	return &rpcErr{msg: msg}
}

type rpcErr struct{ msg string }

func (e *rpcErr) Error() string { return e.msg }

// TestCodexHandshakeStdinHistory asserts that all four daemon-driven
// frames (initialize, initialized, thread/start, turn/start) actually
// landed in the process's stdin, in order.
func TestCodexHandshakeStdinHistory(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	proc, _ := fake.Start(context.Background(), process.Command{Executable: "codex"})
	running := proc.(*process.FakeRunning)

	mustWriteJSONRPC(t, running, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize"})
	mustWriteJSONRPC(t, running, map[string]any{"jsonrpc": "2.0", "method": "initialized"})
	mustWriteJSONRPC(t, running, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "thread/start"})
	mustWriteJSONRPC(t, running, map[string]any{"jsonrpc": "2.0", "id": 3, "method": "turn/start"})

	want := []string{"initialize", "initialized", "thread/start", "turn/start"}
	for i, w := range want {
		select {
		case got := <-running.StdinRecv():
			if !strings.Contains(string(got), w) {
				t.Fatalf("stdin frame #%d: want %q got %q", i, w, got)
			}
		case <-time.After(time.Second):
			t.Fatalf("stdin frame #%d (%s) never written", i, w)
		}
	}
}
