package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestOwnedAndSharedFinalOutputSelection(t *testing.T) {
	message := func(id, phase, text string) map[string]any {
		return map[string]any{"id": id, "type": "agentMessage", "phase": phase, "text": text}
	}
	for _, tc := range []struct {
		name  string
		items []any
		want  string
	}{
		{"final", []any{message("1", "commentary", "progress"), message("2", "final_answer", "answer")}, "answer"},
		{"commentary only", []any{message("1", "commentary", "progress")}, ""},
		{"mixed without final", []any{message("1", "", "unknown"), message("2", "commentary", "progress")}, ""},
		{"legacy last message", []any{message("1", "", "progress"), message("2", "", "answer")}, "answer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, io := ownedTurnDriver(t)
			state := agentbridge.State{}
			for _, item := range tc.items {
				events, _, err := d.OnRaw(t.Context(), makeNotification("item/completed", map[string]any{"threadId": "parent", "turnId": "parent-turn", "item": item}), io)
				if err != nil {
					t.Fatal(err)
				}
				for _, event := range events {
					state, _ = agentbridge.Reduce(state, event, nil)
				}
			}
			events, _, err := d.OnRaw(t.Context(), makeNotification("turn/completed", map[string]any{"threadId": "parent", "turn": map[string]any{"id": "parent-turn", "status": "completed"}}), io)
			if err != nil {
				t.Fatal(err)
			}
			for _, event := range events {
				state, _ = agentbridge.Reduce(state, event, nil)
			}
			if state.Result.Output != tc.want {
				t.Fatalf("owned result=%q want=%q", state.Result.Output, tc.want)
			}
			shared, _ := sharedFixture(t)
			state = agentbridge.State{}
			for _, event := range shared.sharedTurnEvents(map[string]any{"id": "shared-turn", "status": "completed", "completedAt": float64(1), "items": tc.items}) {
				state, _ = agentbridge.Reduce(state, event, nil)
			}
			if state.Result.Output != tc.want {
				t.Fatalf("shared result=%q want=%q", state.Result.Output, tc.want)
			}
		})
	}
}

func TestOwnedTurnBuffersUnidentifiedCompletionUntilReply(t *testing.T) {
	d, io := ownedTurnDriver(t)
	d.owned.id = ""
	d.pending[42] = pendingRequest{method: codexMethodTurnStart}
	events, _, err := d.OnRaw(t.Context(), makeNotification("turn/completed", map[string]any{"turn": map[string]any{"status": "completed"}}), io)
	if err != nil || len(events) != 0 {
		t.Fatalf("unidentified early completion escaped: %v %+v", err, events)
	}
	events, _, err = d.OnRaw(t.Context(), makeResponse(42, map[string]any{"turn": map[string]any{"id": "parent-turn"}}), io)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Kind == agentbridge.EventResult {
			t.Fatalf("unidentified buffered completion escaped: %+v", event)
		}
	}
}

func TestOwnedTurnBoundsEarlyNotifications(t *testing.T) {
	d, io := ownedTurnDriver(t)
	d.owned.id = ""
	d.pending[42] = pendingRequest{method: codexMethodTurnStart}
	d.owned.early = make([]agentbridge.RawEvent, 256)
	events, _, err := d.OnRaw(t.Context(), makeNotification("turn/completed", map[string]any{"threadId": "parent", "turn": map[string]any{"id": "parent-turn", "status": "completed"}}), io)
	if err != nil || len(events) == 0 || events[len(events)-1].Result.Status != agentbridge.ResultFailed {
		t.Fatalf("buffer overflow=%v %+v", err, events)
	}
}
