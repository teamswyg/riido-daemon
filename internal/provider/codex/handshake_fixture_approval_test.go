package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (f *codexHandshakeFixture) approveCommand() {
	approveRequestID := 91
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"id":      approveRequestID,
		"method":  "approve_command",
		"params":  map[string]any{"id": "cmd-1", "command": "ls -la"},
	}))
	if !f.drainUntil(func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		return hasToolApproval(evs, "cmd-1")
	}) {
		f.t.Fatal("approve_command not surfaced as ToolApprovalNeeded")
	}
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0",
		"id":      approveRequestID,
		"result":  map[string]any{"approved": true},
	})
}
