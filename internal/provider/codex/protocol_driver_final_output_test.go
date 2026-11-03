package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexParentFinalMessagePhase(t *testing.T) {
	for _, tc := range []struct {
		name  string
		phase any
		want  agentbridge.ResultStatus
	}{
		{"final answer", "final_answer", agentbridge.ResultCompleted},
		{"older runtime unknown phase", nil, agentbridge.ResultCompleted},
		{"commentary only", "commentary", agentbridge.ResultFailed},
		{"partial answer only", "partial_answer", agentbridge.ResultFailed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, io, state := parentResultFixture(t)
			state = applyResultFrame(t, d, io, state, makeNotification("item/agentMessage/delta", map[string]any{
				"threadId": "parent", "turnId": "parent-turn", "delta": "streamed progress",
			}))
			state = applyResultFrame(t, d, io, state, parentMessage(tc.phase, "completed message"))
			state = applyResultFrame(t, d, io, state, parentCompletion("completed"))
			if state.Result.Status != tc.want {
				t.Fatalf("result=%+v want=%s", state.Result, tc.want)
			}
			if tc.want == agentbridge.ResultCompleted && state.Result.Output != "completed message" {
				t.Fatalf("progress replaced final message: %+v", state.Result)
			}
		})
	}
}

func TestCodexParentFinalAnswerOutranksUnknownPhase(t *testing.T) {
	d, io, state := parentResultFixture(t)
	state = applyResultFrame(t, d, io, state, parentMessage("final_answer", "final answer"))
	state = applyResultFrame(t, d, io, state, parentMessage(nil, "unclassified text"))
	state = applyResultFrame(t, d, io, state, parentCompletion("completed"))
	if state.Result.Output != "final answer" {
		t.Fatalf("explicit final answer lost: %+v", state.Result)
	}
}

func TestCodexParentAwaitsTurnResultAfterThreadStatus(t *testing.T) {
	for _, status := range []string{"idle", "ready", "completed", "finished"} {
		t.Run(status, func(t *testing.T) {
			d, io, state := parentResultFixture(t)
			state = applyResultFrame(t, d, io, state, parentMessage("final_answer", "final answer"))
			state = applyResultFrame(t, d, io, state, makeNotification("thread/status/changed", map[string]any{
				"threadId": "parent", "status": status,
			}))
			if state.Terminal {
				t.Fatal("thread availability completed the parent turn")
			}
			state = applyResultFrame(t, d, io, state, parentCompletion("completed"))
			if state.Result.Status != agentbridge.ResultCompleted || state.Result.Output != "final answer" {
				t.Fatalf("parent result missing: %+v", state.Result)
			}
		})
	}
}

func TestCodexParentProgressBeforeErrorCannotBecomeResult(t *testing.T) {
	d, io, state := parentResultFixture(t)
	state = applyResultFrame(t, d, io, state, makeNotification("item/agentMessage/delta", map[string]any{
		"threadId": "parent", "turnId": "parent-turn", "delta": "progress before error",
	}))
	state = applyResultFrame(t, d, io, state, makeNotification("error", map[string]any{
		"threadId": "parent", "turnId": "parent-turn", "message": "stream disconnected",
	}))
	state = applyResultFrame(t, d, io, state, parentCompletion("completed"))
	if state.Result.Status != agentbridge.ResultFailed || state.Result.Error == "" {
		t.Fatalf("progress became a successful result after error: %+v", state.Result)
	}
}
