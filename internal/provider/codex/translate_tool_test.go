package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateCommandExecutionLifecycle(t *testing.T) {
	start := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"command_execution_started","params":{"id":"c1","command":"ls"}}`))
	if len(start) != 1 || start[0].Kind != agentbridge.EventToolCallStarted {
		t.Fatalf("start: %+v", start)
	}
	if start[0].Tool.ID != "c1" || start[0].Tool.Kind != "shell" || start[0].Tool.Args["command"] != "ls" {
		t.Fatalf("tool ref: %+v", start[0].Tool)
	}

	delta := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"command_execution_output","params":{"id":"c1","chunk":"line1"}}`))
	if len(delta) != 1 || delta[0].Kind != agentbridge.EventToolCallDelta {
		t.Fatalf("delta: %+v", delta)
	}
	done := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"command_execution_completed","params":{"id":"c1","exit_code":0}}`))
	if len(done) != 1 || done[0].Kind != agentbridge.EventToolCallCompleted {
		t.Fatalf("done: %+v", done)
	}
	fail := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"command_execution_completed","params":{"id":"c2","exit_code":2}}`))
	if len(fail) != 1 || fail[0].Kind != agentbridge.EventToolCallFailed {
		t.Fatalf("fail: %+v", fail)
	}
}

func TestTranslateApplyPatch(t *testing.T) {
	start := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"apply_patch_started","params":{"id":"p1"}}`))
	if len(start) != 1 || start[0].Kind != agentbridge.EventToolCallStarted || start[0].Tool.Kind != "patch_apply" {
		t.Fatalf("patch start: %+v", start)
	}
	done := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"apply_patch_completed","params":{"id":"p1"}}`))
	if len(done) != 1 || done[0].Kind != agentbridge.EventToolCallCompleted {
		t.Fatalf("patch done: %+v", done)
	}
}
