package codex

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverErrorNotificationThenEmptyCompletionFails(t *testing.T) {
	d, io := startedProtocolDriver(t)

	events, _, err := d.OnRaw(
		context.Background(),
		makeNotification("error", map[string]any{"message": "failed to connect to websocket: 401 Unauthorized"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != agentbridge.EventError {
		t.Fatalf("expected non-terminal error event first, got %+v", events)
	}

	events, _, err = d.OnRaw(context.Background(), makeNotification("turn_completed", map[string]any{"output": ""}), io)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("expected failed result after empty completion, got %+v", events)
	}
	if !strings.Contains(last.Result.Error, "401 Unauthorized") {
		t.Fatalf("expected retained runtime error, got %+v", last.Result)
	}
}
