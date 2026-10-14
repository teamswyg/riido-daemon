package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverErrorNotificationCanRecoverWithText(t *testing.T) {
	d, io := startedProtocolDriver(t)

	_, _, _ = d.OnRaw(
		context.Background(),
		makeNotification("error", map[string]any{"message": "temporary stream disconnect"}),
		io,
	)
	events, _, err := d.OnRaw(
		context.Background(),
		makeNotification("agent_message", map[string]any{"text": "작업 완료"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta {
		t.Fatalf("expected text delta, got %+v", events)
	}

	events, _, err = d.OnRaw(context.Background(), makeNotification("turn_completed", map[string]any{"output": ""}), io)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("expected recovered completion, got %+v", events)
	}
}
