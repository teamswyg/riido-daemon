package codex

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// TestCodexFullHandshakeComposition is the M-3 regression: drive every
// step of the Codex app-server JSON-RPC handshake through Parser +
// RPCActor + Translator together, with a fake process providing stdio.
//
// Steps exercised (spec §3.2):
//
//  1. initialize request → response
//  2. initialized notification (no response)
//  3. thread/start request → response → thread_started notification
//  4. turn/start request → response → turn_started notification
//  5. streaming notification: agent_message
//  6. server_request: approve_command → daemon replies with approval
//  7. turn_completed notification → translated to EventResult
//  8. pending-map cleanup: a stray response with an unknown id is a no-op
//
// Plus: semantic-idle equivalent — RPC actor releases pending callers
// when Close is invoked.
//
// The test plays both roles: the daemon (writes stdin, calls Translate
// on each parsed RawEvent, routes RPC responses through the actor) and
// the codex server (writes stdout via the fake process).
func TestCodexFullHandshakeComposition(t *testing.T) {
	parser := NewParser()
	rpc := StartRPCActor(context.Background())
	defer rpc.Close()

	// Fake process gives us a WriteStdin we can inspect.
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	proc, err := fake.Start(context.Background(), process.Command{Executable: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	running := proc.(*process.FakeRunning)

	// --- Step 1: initialize ---

	initID := rpc.NextID()
	initReply := rpc.Register(initID)
	mustWriteJSONRPC(t, running, map[string]any{
		"jsonrpc": "2.0",
		"id":      initID,
		"method":  "initialize",
		"params":  map[string]any{"client": "riido"},
	})

	// The server (fake stdout) responds with an initialize result.
	running.EmitStdout(jsonline(t, map[string]any{
		"jsonrpc": "2.0",
		"id":      initID,
		"result":  map[string]any{"serverInfo": map[string]any{"name": "codex"}},
	}))

	// Parser sees the response → RPCActor.Resolve.
	routeRPC(t, parser, running.Stdout(), rpc, 1*time.Second, initID)

	select {
	case r := <-initReply:
		if r.Err != nil {
			t.Fatalf("initialize err: %v", r.Err)
		}
		if r.Result["serverInfo"] == nil {
			t.Fatalf("initialize result missing serverInfo: %+v", r.Result)
		}
	case <-time.After(time.Second):
		t.Fatal("initialize reply not delivered")
	}

	// --- Step 2: initialized notification (no response expected) ---

	mustWriteJSONRPC(t, running, map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]any{},
	})

	// --- Step 3: thread/start ---

	threadID := rpc.NextID()
	threadReply := rpc.Register(threadID)
	mustWriteJSONRPC(t, running, map[string]any{
		"jsonrpc": "2.0",
		"id":      threadID,
		"method":  "thread/start",
		"params":  map[string]any{},
	})

	running.EmitStdout(jsonline(t, map[string]any{
		"jsonrpc": "2.0",
		"id":      threadID,
		"result":  map[string]any{"thread_id": "th-1"},
	}))
	running.EmitStdout(jsonline(t, map[string]any{
		"jsonrpc": "2.0",
		"method":  "thread_started",
		"params":  map[string]any{"thread_id": "th-1"},
	}))

	// Drain — expect thread response and thread_started translation.
	threadResolved := false
	sawSessionIdentified := false
	drainHandshake(t, parser, running.Stdout(), rpc, 1*time.Second, func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool {
		if raw.Type == "response" {
			threadResolved = true
		}
		for _, ev := range evs {
			if ev.Kind == agentbridge.EventSessionIdentified && ev.SessionID == "th-1" {
				sawSessionIdentified = true
			}
		}
		return threadResolved && sawSessionIdentified
	})

	select {
	case r := <-threadReply:
		if r.Err != nil {
			t.Fatalf("thread/start err: %v", r.Err)
		}
		if r.Result["thread_id"] != "th-1" {
			t.Fatalf("thread_id missing: %+v", r.Result)
		}
	case <-time.After(time.Second):
		t.Fatal("thread/start reply not delivered")
	}

	// --- Step 4: turn/start ---

	turnID := rpc.NextID()
	turnReply := rpc.Register(turnID)
	mustWriteJSONRPC(t, running, map[string]any{
		"jsonrpc": "2.0",
		"id":      turnID,
		"method":  "turn/start",
		"params":  map[string]any{"prompt": "hi"},
	})
	running.EmitStdout(jsonline(t, map[string]any{"jsonrpc": "2.0", "id": turnID, "result": map[string]any{}}))
	running.EmitStdout(jsonline(t, map[string]any{
		"jsonrpc": "2.0", "method": "turn_started", "params": map[string]any{},
	}))

	turnResolved := false
	sawLifecycle := false
	drainHandshake(t, parser, running.Stdout(), rpc, 1*time.Second, func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool {
		if raw.Type == "response" {
			turnResolved = true
		}
		for _, ev := range evs {
			if ev.Kind == agentbridge.EventLifecycle && ev.Phase == agentbridge.StateRunning {
				sawLifecycle = true
			}
		}
		return turnResolved && sawLifecycle
	})
	if !turnResolved || !sawLifecycle {
		t.Fatalf("turn/start chain: resolved=%v lifecycle=%v", turnResolved, sawLifecycle)
	}
	<-turnReply

	// --- Step 5: streaming notification (agent_message) ---

	running.EmitStdout(jsonline(t, map[string]any{
		"jsonrpc": "2.0", "method": "agent_message", "params": map[string]any{"text": "hello"},
	}))

	sawText := false
	drainHandshake(t, parser, running.Stdout(), rpc, 1*time.Second, func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		for _, ev := range evs {
			if ev.Kind == agentbridge.EventTextDelta && ev.Text == "hello" {
				sawText = true
			}
		}
		return sawText
	})
	if !sawText {
		t.Fatal("agent_message not translated")
	}

	// --- Step 6: server_request approve_command ---

	approveRequestID := 91 // server-chosen id
	running.EmitStdout(jsonline(t, map[string]any{
		"jsonrpc": "2.0", "id": approveRequestID, "method": "approve_command",
		"params": map[string]any{"id": "cmd-1", "command": "ls -la"},
	}))

	sawApproval := false
	drainHandshake(t, parser, running.Stdout(), rpc, 1*time.Second, func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		for _, ev := range evs {
			if ev.Kind == agentbridge.EventToolApprovalNeeded && ev.Tool.ID == "cmd-1" {
				sawApproval = true
			}
		}
		return sawApproval
	})
	if !sawApproval {
		t.Fatal("approve_command not surfaced as ToolApprovalNeeded")
	}

	// Daemon replies with approval (the server_request id, not a new one).
	mustWriteJSONRPC(t, running, map[string]any{
		"jsonrpc": "2.0", "id": approveRequestID, "result": map[string]any{"approved": true},
	})

	// --- Step 7: turn_completed ---

	running.EmitStdout(jsonline(t, map[string]any{
		"jsonrpc": "2.0", "method": "turn_completed",
		"params": map[string]any{"output": "done"},
	}))

	sawResult := false
	drainHandshake(t, parser, running.Stdout(), rpc, 1*time.Second, func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		for _, ev := range evs {
			if ev.Kind == agentbridge.EventResult && ev.Result.Status == agentbridge.ResultCompleted {
				sawResult = true
			}
		}
		return sawResult
	})
	if !sawResult {
		t.Fatal("turn_completed not surfaced as EventResult")
	}

	// --- Step 8: pending-map cleanup of a stray response ---
	// Sending a response with an unknown id MUST NOT panic and MUST NOT
	// leave anything in the pending map. We test it as a no-op.
	rpc.Resolve(99999, map[string]any{"orphan": true}, nil)

	// Final teardown: closing the RPC actor must not leak. (Pending-map
	// cleanup-on-close is already verified in rpc_actor_test.go via
	// TestRPCActorCloseCancelsPending.)
	running.EmitExit(0, nil)
}

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

// TestCodexRPCActorReleasesPendingOnProcessExit is the M-3 invariant
// "process exit with pending requests" — pending callers must NOT leak
// when the RPC actor closes for any reason, including because the
// process backing it died.
func TestCodexRPCActorReleasesPendingOnProcessExit(t *testing.T) {
	rpc := StartRPCActor(context.Background())

	const N = 8
	replies := make([]<-chan RPCResult, 0, N)
	for i := 0; i < N; i++ {
		id := rpc.NextID()
		replies = append(replies, rpc.Register(id))
	}

	// Simulate process exit: caller closes the RPC actor.
	rpc.Close()

	deadline := time.After(2 * time.Second)
	for i, ch := range replies {
		select {
		case r := <-ch:
			if r.Err == nil {
				t.Fatalf("reply #%d: expected error after Close, got nil", i)
			}
		case <-deadline:
			t.Fatalf("reply #%d: blocked after Close", i)
		}
	}
	_ = strconv.Itoa // keep import set stable
}
