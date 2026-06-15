package codex

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type codexHandshakeFixture struct {
	t       *testing.T
	parser  agentbridge.Parser
	rpc     *RPCActor
	running *process.FakeRunning
}

func newCodexHandshakeFixture(t *testing.T) *codexHandshakeFixture {
	t.Helper()
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	proc, err := fake.Start(context.Background(), process.Command{Executable: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	return &codexHandshakeFixture{
		t:       t,
		parser:  NewParser(),
		rpc:     StartRPCActor(context.Background()),
		running: proc.(*process.FakeRunning),
	}
}

func (f *codexHandshakeFixture) close() {
	f.rpc.Close()
}

func (f *codexHandshakeFixture) initialize() {
	initID := f.rpc.NextID()
	initReply := f.rpc.Register(initID)
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0",
		"id":      initID,
		"method":  "initialize",
		"params":  map[string]any{"client": "riido"},
	})
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"id":      initID,
		"result":  map[string]any{"serverInfo": map[string]any{"name": "codex"}},
	}))
	routeRPC(f.t, f.parser, f.running.Stdout(), f.rpc, time.Second, initID)
	select {
	case result := <-initReply:
		if result.Err != nil {
			f.t.Fatalf("initialize err: %v", result.Err)
		}
		if result.Result["serverInfo"] == nil {
			f.t.Fatalf("initialize result missing serverInfo: %+v", result.Result)
		}
	case <-time.After(time.Second):
		f.t.Fatal("initialize reply not delivered")
	}
}

func (f *codexHandshakeFixture) notifyInitialized() {
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]any{},
	})
}

func (f *codexHandshakeFixture) startThread() {
	threadID := f.rpc.NextID()
	threadReply := f.rpc.Register(threadID)
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0",
		"id":      threadID,
		"method":  "thread/start",
		"params":  map[string]any{},
	})
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"id":      threadID,
		"result":  map[string]any{"thread_id": "th-1"},
	}))
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"method":  "thread_started",
		"params":  map[string]any{"thread_id": "th-1"},
	}))
	f.drainUntil(func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool {
		return raw.Type == "response" && hasSessionIdentified(evs, "th-1")
	})
	select {
	case result := <-threadReply:
		if result.Err != nil {
			f.t.Fatalf("thread/start err: %v", result.Err)
		}
		if result.Result["thread_id"] != "th-1" {
			f.t.Fatalf("thread_id missing: %+v", result.Result)
		}
	case <-time.After(time.Second):
		f.t.Fatal("thread/start reply not delivered")
	}
}

func (f *codexHandshakeFixture) startTurn() {
	turnID := f.rpc.NextID()
	turnReply := f.rpc.Register(turnID)
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0",
		"id":      turnID,
		"method":  "turn/start",
		"params":  map[string]any{"prompt": "hi"},
	})
	f.running.EmitStdout(jsonline(f.t, map[string]any{"jsonrpc": "2.0", "id": turnID, "result": map[string]any{}}))
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0", "method": "turn_started", "params": map[string]any{},
	}))
	turnResolved, sawLifecycle := false, false
	f.drainUntil(func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool {
		turnResolved = turnResolved || raw.Type == "response"
		sawLifecycle = sawLifecycle || hasLifecycleRunning(evs)
		return turnResolved && sawLifecycle
	})
	if !turnResolved || !sawLifecycle {
		f.t.Fatalf("turn/start chain: resolved=%v lifecycle=%v", turnResolved, sawLifecycle)
	}
	<-turnReply
}

func (f *codexHandshakeFixture) streamAgentMessage() {
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0", "method": "agent_message", "params": map[string]any{"text": "hello"},
	}))
	if !f.drainUntil(func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		return hasTextDelta(evs, "hello")
	}) {
		f.t.Fatal("agent_message not translated")
	}
}

func (f *codexHandshakeFixture) approveCommand() {
	approveRequestID := 91
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0", "id": approveRequestID, "method": "approve_command",
		"params": map[string]any{"id": "cmd-1", "command": "ls -la"},
	}))
	if !f.drainUntil(func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		return hasToolApproval(evs, "cmd-1")
	}) {
		f.t.Fatal("approve_command not surfaced as ToolApprovalNeeded")
	}
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0", "id": approveRequestID, "result": map[string]any{"approved": true},
	})
}

func (f *codexHandshakeFixture) completeTurn() {
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0", "method": "turn_completed",
		"params": map[string]any{"output": "done"},
	}))
	if !f.drainUntil(func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		return hasCompletedResult(evs)
	}) {
		f.t.Fatal("turn_completed not surfaced as EventResult")
	}
}

func (f *codexHandshakeFixture) resolveOrphanResponse() {
	f.rpc.Resolve(99999, map[string]any{"orphan": true}, nil)
}

func (f *codexHandshakeFixture) finish() {
	f.running.EmitExit(0, nil)
}

func (f *codexHandshakeFixture) drainUntil(done func(agentbridge.RawEvent, []agentbridge.Event) bool) bool {
	reached := false
	drainHandshake(f.t, f.parser, f.running.Stdout(), f.rpc, time.Second, func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool {
		reached = done(raw, evs)
		return reached
	})
	return reached
}
