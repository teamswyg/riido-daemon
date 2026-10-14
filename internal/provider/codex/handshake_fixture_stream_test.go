package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (f *codexHandshakeFixture) streamAgentMessage() {
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"method":  "agent_message",
		"params":  map[string]any{"text": "hello"},
	}))
	if !f.drainUntil(func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		return hasTextDelta(evs, "hello")
	}) {
		f.t.Fatal("agent_message not translated")
	}
}

func (f *codexHandshakeFixture) completeTurn() {
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"method":  "turn_completed",
		"params":  map[string]any{"output": "done"},
	}))
	if !f.drainUntil(func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		return hasCompletedResult(evs)
	}) {
		f.t.Fatal("turn_completed not surfaced as EventResult")
	}
}
