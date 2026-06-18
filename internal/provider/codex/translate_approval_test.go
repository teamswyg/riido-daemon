package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateApprovalRequest(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":7,"method":"approve_command","params":{"command":"rm -rf /","id":"cmd-7"}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("approval: %+v", evs)
	}
	if evs[0].Tool.Args["command"] != "rm -rf /" {
		t.Fatalf("approval args: %+v", evs[0].Tool.Args)
	}
}

func TestTranslatePatchApprovalCapturesPathArg(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":8,"method":"approve_patch","params":{"path":".git/config","id":"patch-8"}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("approval: %+v", evs)
	}
	if evs[0].Tool.Args["path"] != ".git/config" {
		t.Fatalf("patch args: %+v", evs[0].Tool.Args)
	}
}
