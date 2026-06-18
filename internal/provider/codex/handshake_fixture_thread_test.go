package codex

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
