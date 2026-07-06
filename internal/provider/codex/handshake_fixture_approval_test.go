package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (f *codexHandshakeFixture) approveCommand() {
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"id":      91,
		"method":  "approve_command",
		"params":  map[string]any{"id": "cmd-1", "command": "ls -la"},
	}))
	if !f.drainUntil(func(_ agentbridge.RawEvent, evs []agentbridge.Event) bool {
		return hasAutoApprovalLog(evs)
	}) {
		f.t.Fatal("approve_command not auto-approved")
	}
}
