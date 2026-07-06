package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateApprovalRequest(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":7,"method":"approve_command","params":{"command":"rm -rf /","id":"cmd-7"}}`)
	evs, cmds := txFull(t, raw)
	assertAutoApproval(t, evs, cmds, "7")
}

func TestTranslatePatchApprovalCapturesPathArg(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":8,"method":"approve_patch","params":{"path":".git/config","id":"patch-8"}}`)
	evs, cmds := txFull(t, raw)
	assertAutoApproval(t, evs, cmds, "8")
}

func TestTranslateFileChangeRequestApproval(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":77,"method":"item/fileChange/requestApproval","params":{"threadId":"th-1","turnId":"turn-1","itemId":"item-1","startedAtMs":1780000000000,"reason":"write file","grantRoot":"/tmp/work"}}`)
	evs, cmds := txFull(t, raw)
	assertAutoApproval(t, evs, cmds, "77")
}

func TestTranslateCommandExecutionRequestApproval(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":78,"method":"item/commandExecution/requestApproval","params":{"threadId":"th-1","turnId":"turn-1","itemId":"item-2","approvalId":"appr-1","startedAtMs":1780000000000,"command":"go test ./...","cwd":"/tmp/work"}}`)
	evs, cmds := txFull(t, raw)
	assertAutoApproval(t, evs, cmds, "78")
}

func assertAutoApproval(t *testing.T, evs []agentbridge.Event, cmds []agentbridge.Command, reqID string) {
	t.Helper()
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventLog {
		t.Fatalf("auto approval event: %+v", evs)
	}
	if len(cmds) != 1 || cmds[0].Kind != agentbridge.CommandApproveTool {
		t.Fatalf("auto approval command: %+v", cmds)
	}
	if cmds[0].ProviderRequestID != reqID {
		t.Fatalf("provider request id = %q, want %q", cmds[0].ProviderRequestID, reqID)
	}
}
