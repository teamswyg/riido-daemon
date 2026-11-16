package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestOwnedCommandOutputIsPrivateCorrelatedActivity(t *testing.T) {
	for _, tc := range []struct {
		name, thread, turn, item, delta string
		activity                        bool
	}{
		{"owned output", "parent", "parent-turn", "command", "private output", true},
		{"other thread", "child", "parent-turn", "command", "private output", false},
		{"other turn", "parent", "older-turn", "command", "private output", false},
		{"missing thread", "", "parent-turn", "command", "private output", false},
		{"missing turn", "parent", "", "command", "private output", false},
		{"missing item", "parent", "parent-turn", "", "private output", false},
		{"empty output", "parent", "parent-turn", "command", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, io := ownedTurnDriver(t)
			events, _, err := d.OnRaw(context.Background(), makeNotification("item/commandExecution/outputDelta", map[string]any{
				"threadId": tc.thread, "turnId": tc.turn, "itemId": tc.item, "delta": tc.delta,
			}), io)
			if err != nil {
				t.Fatal(err)
			}
			activity := false
			for _, event := range events {
				activity = activity || event.Kind == agentbridge.EventSemanticActivity
				if event.Text == tc.delta && tc.delta != "" {
					t.Fatalf("command output leaked: %+v", event)
				}
			}
			if activity != tc.activity {
				t.Fatalf("activity = %v, want %v; events=%+v", activity, tc.activity, events)
			}
		})
	}
}
