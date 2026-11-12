package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Exercise the real handshake and reducer: another thread's completion must
// not terminate the parent while its review/posting work remains in flight.
func TestOwnedTurnIgnoresChildCompletion(t *testing.T) {
	d, io := ownedTurnDriver(t)
	for _, raw := range []agentbridge.RawEvent{
		makeNotification("thread/started", map[string]any{"thread": map[string]any{"id": "child"}}),
		makeNotification("turn/started", map[string]any{"threadId": "child", "turn": map[string]any{"id": "child-turn"}}),
		makeNotification("item/agentMessage/delta", map[string]any{"threadId": "child", "turnId": "child-turn", "delta": "child review"}),
		makeNotification("turn/completed", map[string]any{"threadId": "child", "turn": map[string]any{"id": "child-turn", "status": "completed"}}),
	} {
		events, _, err := d.OnRaw(context.Background(), raw, io)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			if event.Kind == agentbridge.EventResult || event.Kind == agentbridge.EventTextDelta || event.Kind == agentbridge.EventSessionIdentified {
				t.Errorf("child event %s reached parent: %+v", raw.Type, event)
			}
		}
	}
}

func TestOwnedTurnFinalAnswerExcludesCommentary(t *testing.T) {
	d, io := ownedTurnDriver(t)
	state := agentbridge.State{}
	for _, raw := range []agentbridge.RawEvent{
		makeNotification("item/started", map[string]any{"threadId": "parent", "turnId": "parent-turn", "item": map[string]any{"id": "progress", "type": "agentMessage", "phase": "commentary", "text": ""}}),
		makeNotification("item/agentMessage/delta", map[string]any{"threadId": "parent", "turnId": "parent-turn", "itemId": "progress", "delta": "Review in progress."}),
		makeNotification("item/completed", map[string]any{"threadId": "parent", "turnId": "parent-turn", "item": map[string]any{"id": "progress", "type": "agentMessage", "phase": "commentary", "text": "Review in progress."}}),
		makeNotification("item/completed", map[string]any{"threadId": "parent", "turnId": "parent-turn", "item": map[string]any{"id": "answer", "type": "agentMessage", "phase": "final_answer", "text": "Review posted: https://example.test/review/1"}}),
		makeNotification("turn/completed", map[string]any{"threadId": "parent", "turn": map[string]any{"id": "parent-turn", "status": "completed"}}),
	} {
		events, _, err := d.OnRaw(context.Background(), raw, io)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			state, _ = agentbridge.Reduce(state, event, nil)
		}
	}
	if state.Result.Status != agentbridge.ResultCompleted || state.Result.Output != "Review posted: https://example.test/review/1" {
		t.Fatalf("parent final result: %+v", state.Result)
	}
}

func ownedTurnDriver(t *testing.T) (*protocolDriver, *recordingIO) {
	t.Helper()
	driver, _ := NewProtocolDriver(agentbridge.StartRequest{Prompt: "Review and post results"})
	d := driver.(*protocolDriver)
	io := newRecordingIO()
	if err := d.OnStart(context.Background(), io); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []agentbridge.RawEvent{
		makeResponse(1, map[string]any{}),
		makeResponse(2, map[string]any{"thread": map[string]any{"id": "parent"}}),
		makeResponse(3, map[string]any{"turn": map[string]any{"id": "parent-turn", "status": "inProgress"}}),
	} {
		if _, _, err := d.OnRaw(context.Background(), raw, io); err != nil {
			t.Fatal(err)
		}
	}
	return d, io
}
