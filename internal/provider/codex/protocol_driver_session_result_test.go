package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexParentResultExcludesChildAndProgress(t *testing.T) {
	d, io, state := parentResultFixture(t)
	state = applyResultFrame(t, d, io, state, makeNotification("item/agentMessage/delta", map[string]any{
		"threadId": "parent", "turnId": "parent-turn", "delta": "검토 중입니다",
	}))
	for _, method := range []string{"item/agentMessage/delta", "item/completed", "turn/completed"} {
		state = applyResultFrame(t, d, io, state, makeNotification(method, map[string]any{
			"threadId": "child", "turnId": "child-turn", "delta": "지적 없음",
			"item": map[string]any{"type": "agentMessage", "phase": "final_answer", "text": "지적 없음"},
			"turn": map[string]any{"id": "child-turn", "status": "completed"},
		}))
	}
	if state.Terminal || string(state.Output) != "검토 중입니다" {
		t.Fatalf("child changed parent state: %+v", state)
	}
	state = applyResultFrame(t, d, io, state, parentMessage("final_answer", "전체 검토 결과입니다"))
	if state.Terminal {
		t.Fatal("a final message item alone completed the parent turn")
	}
	state = applyResultFrame(t, d, io, state, parentCompletion("completed"))
	if !state.Terminal || state.Result.Status != agentbridge.ResultCompleted || state.Result.Output != "전체 검토 결과입니다" {
		t.Fatalf("wrong parent result: %+v", state.Result)
	}
}

func TestCodexParentResultPreservesNativeFailure(t *testing.T) {
	for _, status := range []string{"failed", "interrupted"} {
		t.Run(status, func(t *testing.T) {
			d, io, state := parentResultFixture(t)
			state = applyResultFrame(t, d, io, state, parentMessage("final_answer", "a partial review"))
			state = applyResultFrame(t, d, io, state, parentCompletion(status))
			want := agentbridge.ResultFailed
			if status == "interrupted" {
				want = agentbridge.ResultCancelled
			}
			if !state.Terminal || state.Result.Status != want {
				t.Fatalf("final text hid native status: %+v", state.Result)
			}
		})
	}
}
