package codex

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (f *codexHandshakeFixture) startTurn() {
	turnID := f.rpc.NextID()
	turnReply := f.rpc.Register(turnID)
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0",
		"id":      turnID,
		"method":  "turn/start",
		"params":  map[string]any{"prompt": "hi"},
	})
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"id":      turnID,
		"result":  map[string]any{},
	}))
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"method":  "turn_started",
		"params":  map[string]any{},
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
	select {
	case result := <-turnReply:
		if result.Err != nil {
			f.t.Fatalf("turn/start err: %v", result.Err)
		}
	case <-time.After(time.Second):
		f.t.Fatal("turn/start reply not delivered")
	}
}
