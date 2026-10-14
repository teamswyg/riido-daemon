package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverApprovalRequestTranslatesToApprovalEvent(t *testing.T) {
	d, io := startedProtocolDriver(t)

	events, _, err := d.OnRaw(
		context.Background(),
		makeServerRequest(42, "approve_command", map[string]any{"id": "cmd-1", "command": "ls"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("approval: %+v", events)
	}
	if events[0].Tool.ID != "cmd-1" {
		t.Fatalf("tool ref: %+v", events[0].Tool)
	}
}
