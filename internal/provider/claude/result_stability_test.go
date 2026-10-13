package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNativeClaudeFailuresKeepReasonAndNeverComplete(t *testing.T) {
	tests := []struct {
		frame  string
		status agentbridge.ResultStatus
		reason string
	}{
		{`{"type":"result","subtype":"error_during_execution","is_error":true,"errors":["Not logged in","Run /login"]}`, agentbridge.ResultFailed, "Not logged in\nRun /login"},
		{`{"type":"result","subtype":"success","is_error":true,"result":"Invalid API key"}`, agentbridge.ResultFailed, "Invalid API key"},
		{`{"type":"result","subtype":"error_max_turns","errors":["Turn limit reached"]}`, agentbridge.ResultBlocked, "Turn limit reached"},
		{`{"type":"result","subtype":"error_max_budget_usd","errors":["Budget exceeded"]}`, agentbridge.ResultBlocked, "Budget exceeded"},
		{`{"type":"result","subtype":"error_future_condition"}`, agentbridge.ResultFailed, "claude result: error_future_condition"},
	}
	for _, tt := range tests {
		events, _, err := Translate(mustParseRaw(t, tt.frame))
		if err != nil {
			t.Fatal(err)
		}
		result := events[len(events)-1].Result
		if result.Status != tt.status || result.Error != tt.reason {
			t.Fatalf("%s: %+v", tt.frame, result)
		}
	}
}
