package codex

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestOwnedIdleContinuationIsBoundedAndCorrelated(t *testing.T) {
	for _, tc := range []struct{ name, thread, turn, status string }{
		{"still running twice", "parent", "parent-turn", "inProgress"},
		{"foreign thread", "child", "parent-turn", "inProgress"},
		{"previous turn", "parent", "previous-turn", "inProgress"},
		{"missing turn", "parent", "", "inProgress"},
		{"unknown status", "parent", "parent-turn", ""},
		{"failed turn", "parent", "parent-turn", "failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				sess, child := startOwnedSessionWithIdle(t, waitIdle)
				id := receiveIdleRead(t, child)
				child.EmitStdout(idleReadReply(t, id, tc.thread, tc.turn, tc.status))
				if tc.name == "still running twice" {
					id = receiveIdleRead(t, child)
					child.EmitStdout(idleReadReply(t, id, tc.thread, tc.turn, tc.status))
				}
				if result := <-sess.Result(); result.Status != agentbridge.ResultTimeout || result.Error != "semantic idle timeout" {
					t.Fatalf("result = %+v", result)
				}
			})
		})
	}
}

func TestOwnedIdleGraceKeepsHardDeadlineAndCancellation(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		synctest.Test(t, func(t *testing.T) {
			sess, child := startOwnedSessionWithIdle(t, 3*time.Second)
			id := receiveIdleRead(t, child)
			child.EmitStdout(idleReadReply(t, id, "parent", "parent-turn", "inProgress"))
			synctest.Wait()
			want := agentbridge.ResultTimeout
			if cancel {
				sess.Cancel(nil)
				want = agentbridge.ResultCancelled
			}
			select {
			case result := <-sess.Result():
				if result.Status != want {
					t.Fatalf("result = %+v", result)
				}
				// The hard timer and its context deadline expire together. Either
				// existing error spelling is valid; the original deadline is exact.
				if !cancel && (result.Error != "hard timeout" && result.Error != "run hard timeout exceeded (5s)" || result.FinishedAt.Sub(result.StartedAt) != 5*time.Second) {
					t.Fatalf("hard deadline changed: %+v", result)
				}
			case <-time.After(2100 * time.Millisecond):
				t.Fatal("idle continuation extended the hard deadline")
			}
		})
	}
}

func TestOwnedProgressDuringIdleReadIgnoresLateReply(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sess, child := startOwnedSessionWithIdle(t, waitIdle)
		id := receiveIdleRead(t, child)
		emitOwnedWaitNotification(child)("item/agentMessage/delta", map[string]any{
			"threadId": "parent", "turnId": "parent-turn", "itemId": "progress", "delta": "working again",
		})
		// A stale, empty read must not terminate the recovered stream.
		child.EmitStdout(idleReadReply(t, id, "parent", "parent-turn", ""))
		synctest.Wait()
		emitIdleCompletion(t, child)
		if result := <-sess.Result(); result.Status != agentbridge.ResultCompleted {
			t.Fatalf("result = %+v", result)
		}
	})
}
