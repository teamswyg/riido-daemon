package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionPropagatesSessionID(t *testing.T) {
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type != "chunk" {
				return nil, nil, nil
			}
			return []agentbridge.Event{
				{Kind: agentbridge.EventSessionIdentified, SessionID: "sess-abc"},
				{Kind: agentbridge.EventResult, Result: completedResult()},
			}, nil, nil
		},
	}
	scenario := startToolGateScenario(t, "task-5", adapter, nil)
	go scenario.running.EmitStdout([]byte("x"))

	res := waitResult(t, scenario.session, time.Second)
	if res.SessionID != "sess-abc" {
		t.Fatalf("session id: %q", res.SessionID)
	}
}
