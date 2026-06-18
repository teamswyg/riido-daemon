package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverTurnCompletedReturnsResult(t *testing.T) {
	d, io := startedProtocolDriver(t)

	events, _, err := d.OnRaw(
		context.Background(),
		makeNotification("turn_completed", map[string]any{"output": "all done"}),
		io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("no events")
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}
}
