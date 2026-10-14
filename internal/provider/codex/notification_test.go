package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestUserInputNotification(t *testing.T) {
	events, _, err := Translate(agentbridge.RawEvent{Type: "server_request:item/tool/requestUserInput", Payload: map[string]any{"id": 42, "params": map[string]any{"itemId": "question"}}})
	if err != nil || len(events) != 1 || events[0].Phase != agentbridge.StateWaitingUserInput || events[0].Tool.ProviderRequestID != "42" {
		t.Fatalf("input event: %+v, %v", events, err)
	}
	_, commands := agentbridge.Reduce(agentbridge.State{}, events[0], nil)
	if len(commands) != 0 {
		t.Fatal("answer request auto-approved")
	}
}
