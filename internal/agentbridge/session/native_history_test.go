package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNativeHistoryIsBlockedBeforePublicationWhenSessionIsForeign(t *testing.T) {
	adapter := &recordingAdapter{name: "fake", parser: &recordingParser{}, translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
		return []agentbridge.Event{
			{Kind: agentbridge.EventSessionIdentified, SessionID: "expected"},
			{Kind: agentbridge.EventProgress, NativeHistory: &agentbridge.NativeHistoryMessage{SessionID: "foreign", Body: "must not be published"}},
			{Kind: agentbridge.EventResult, Result: completedResult()},
		}, nil, nil
	}}
	scenario := startToolGateScenario(t, "native-history-scope", adapter, func(cfg *Config) { cfg.Request.ResumeSessionID = "expected" })
	go scenario.running.EmitStdout([]byte("x"))
	result := waitResult(t, scenario.session, time.Second)
	if result.Status != agentbridge.ResultBlocked {
		t.Fatalf("foreign history succeeded: %+v", result)
	}
	for ev := range scenario.session.Events() {
		if ev.NativeHistory != nil {
			t.Fatal("foreign transcript reached publisher")
		}
	}
}
