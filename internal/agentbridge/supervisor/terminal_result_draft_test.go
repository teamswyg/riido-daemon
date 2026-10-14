package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTerminalResultDraftMapsTaskTransitions(t *testing.T) {
	for _, tc := range terminalResultDraftCases() {
		t.Run(tc.name, func(t *testing.T) {
			got, payload := terminalResultDraft(tc.res)
			if got != tc.want {
				t.Fatalf("event type = %s, want %s", got, tc.want)
			}
			if len(payload) == 0 {
				t.Fatalf("payload must not be empty")
			}
			if !got.IsTransition() {
				t.Fatalf("%s must be an IR transition event", got)
			}
		})
	}
}

func terminalResultDraftCases() []struct {
	name string
	res  agentbridge.Result
	want ir.EventType
} {
	return []struct {
		name string
		res  agentbridge.Result
		want ir.EventType
	}{
		{"completed", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}, ir.EventRunReportedDone},
		{"failed", agentbridge.Result{Status: agentbridge.ResultFailed, Error: "boom"}, ir.EventTaskFailed},
		{"blocked", agentbridge.Result{Status: agentbridge.ResultBlocked, Error: "capability"}, ir.EventTaskFailed},
		{"aborted", agentbridge.Result{Status: agentbridge.ResultAborted, Error: "exit"}, ir.EventTaskFailed},
		{"cancelled", agentbridge.Result{Status: agentbridge.ResultCancelled, Error: "user"}, ir.EventTaskCancelled},
		{"timeout", agentbridge.Result{Status: agentbridge.ResultTimeout, Error: "semantic idle timeout"}, ir.EventTaskTimedOut},
	}
}
