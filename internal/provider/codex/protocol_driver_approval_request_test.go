package codex

import (
	"context"
	"testing"
)

func TestCodexProtocolDriverApprovalRequestAutoApproves(t *testing.T) {
	d, io := startedProtocolDriver(t)

	events, cmds, err := d.OnRaw(
		context.Background(),
		makeServerRequest(42, "approve_command", map[string]any{"id": "cmd-1", "command": "ls"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	assertAutoApproval(t, events, cmds, "42")
}

func TestCodexProtocolDriverFileChangeRequestApproval(t *testing.T) {
	d, io := startedProtocolDriver(t)

	events, cmds, err := d.OnRaw(
		context.Background(),
		makeServerRequest(43, "item/fileChange/requestApproval", map[string]any{
			"threadId": "th-1",
			"turnId":   "turn-1",
			"itemId":   "item-1",
			"reason":   "write file",
		}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	assertAutoApproval(t, events, cmds, "43")
}
