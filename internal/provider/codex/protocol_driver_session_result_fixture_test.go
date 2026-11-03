package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func parentResultFixture(t *testing.T) (*protocolDriver, *recordingIO, agentbridge.State) {
	t.Helper()
	driver, io := startedProtocolDriver(t)
	d := driver.(*protocolDriver)
	state := agentbridge.NewState()
	for _, raw := range []agentbridge.RawEvent{
		makeResponse(1, nil),
		makeResponse(2, map[string]any{"thread": map[string]any{"id": "parent"}}),
		makeResponse(3, map[string]any{"turn": map[string]any{"id": "parent-turn", "status": "inProgress"}}),
	} {
		state = applyResultFrame(t, d, io, state, raw)
	}
	return d, io, state
}

func applyResultFrame(t *testing.T, d *protocolDriver, io *recordingIO, state agentbridge.State, raw agentbridge.RawEvent) agentbridge.State {
	t.Helper()
	events, _, err := d.OnRaw(context.Background(), raw, io)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		state, _ = agentbridge.Reduce(state, event, nil)
	}
	return state
}

func parentMessage(phase any, text string) agentbridge.RawEvent {
	return makeNotification("item/completed", map[string]any{
		"threadId": "parent", "turnId": "parent-turn",
		"item": map[string]any{"id": "message", "type": "agentMessage", "phase": phase, "text": text},
	})
}

func parentCompletion(status string) agentbridge.RawEvent {
	return makeNotification("turn/completed", map[string]any{
		"threadId": "parent", "turn": map[string]any{"id": "parent-turn", "status": status},
	})
}
