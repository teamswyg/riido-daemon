package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverStreamingNotificationTranslatesToEvents(t *testing.T) {
	d, io := startedProtocolDriver(t)

	events, _, err := d.OnRaw(
		context.Background(),
		makeNotification("agent_message", map[string]any{"text": "hi"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta || events[0].Text != "hi" {
		t.Fatalf("text delta: %+v", events)
	}
}
