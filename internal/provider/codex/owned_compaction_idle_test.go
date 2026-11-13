package codex

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestOwnedCompactionCanFinishAfterSemanticIdle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sess, child := startOwnedSessionWithIdle(t, waitIdle)
		emit := emitOwnedWaitNotification(child)
		item := compactionItem("parent", "parent-turn", "compact")
		emit("item/started", item)
		time.Sleep(waitIdle + 20*time.Millisecond)
		select {
		case frame := <-child.StdinRecv():
			t.Fatalf("watchdog began terminal recovery during active compaction: %s", frame)
		default:
		}
		emit("item/completed", item)
		emit("turn/completed", map[string]any{
			"threadId": "parent",
			"turn": map[string]any{
				"id": "parent-turn", "status": "completed",
				"items": []any{map[string]any{"id": "answer", "type": "agentMessage", "phase": "final_answer", "text": "continued after compaction"}},
			},
		})
		if result := <-sess.Result(); result.Status != agentbridge.ResultCompleted || result.Output != "continued after compaction" {
			t.Fatalf("result = %+v", result)
		}
	})
}

func TestOwnedCompactionIdleGraceIsBoundedAndCorrelated(t *testing.T) {
	for _, tc := range []struct {
		name        string
		thread      string
		turn        string
		completedID string
		deadline    time.Duration
	}{
		{"other thread", "child", "parent-turn", "", waitIdle},
		{"other turn", "parent", "previous-turn", "", waitIdle},
		{"missing turn", "parent", "", "", waitIdle},
		{"finished compaction", "parent", "parent-turn", "compact", waitIdle},
		{"stalled compaction", "parent", "parent-turn", "", 2 * waitIdle},
		{"mismatched completion", "parent", "parent-turn", "other-item", 2 * waitIdle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				sess, child := startOwnedSessionWithIdle(t, waitIdle)
				emit := emitOwnedWaitNotification(child)
				item := compactionItem(tc.thread, tc.turn, "compact")
				emit("item/started", item)
				if tc.completedID != "" {
					emit("item/completed", compactionItem(tc.thread, tc.turn, tc.completedID))
				}
				time.Sleep(tc.deadline - 20*time.Millisecond)
				emit("item/started", item)
				select {
				case frame := <-child.StdinRecv():
					t.Fatalf("watchdog fired early: %s", frame)
				default:
				}
				time.Sleep(40 * time.Millisecond)
				select {
				case wire := <-child.StdinRecv():
					var request map[string]any
					if err := json.Unmarshal(wire, &request); err != nil || request["method"] != "thread/read" {
						t.Fatalf("unexpected recovery request: %s (%v)", wire, err)
					}
					child.EmitStdout(jsonline(t, map[string]any{
						"jsonrpc": "2.0", "id": request["id"],
						"result": map[string]any{"thread": map[string]any{"id": "parent"}},
					}))
				default:
					t.Fatal("watchdog did not recover at the bounded idle deadline")
				}
				if result := <-sess.Result(); result.Status != agentbridge.ResultTimeout || result.Error != "semantic idle timeout" {
					t.Fatalf("result = %+v", result)
				}
			})
		})
	}
}

func TestOwnedCompactionGraceKeepsHardDeadlineAndCancellation(t *testing.T) {
	for _, name := range []string{"hard deadline", "cancel"} {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				sess, child := startOwnedSessionWithIdle(t, 3*time.Second)
				emitOwnedWaitNotification(child)("item/started", compactionItem("parent", "parent-turn", "compact"))
				time.Sleep(4 * time.Second)
				want := agentbridge.ResultTimeout
				if name == "cancel" {
					sess.Cancel(nil)
					want = agentbridge.ResultCancelled
				}
				select {
				case result := <-sess.Result():
					if result.Status != want || (want == agentbridge.ResultTimeout && !strings.Contains(result.Error, "hard timeout")) {
						t.Fatalf("result = %+v", result)
					}
				case <-time.After(1500 * time.Millisecond):
					t.Fatal("compaction grace delayed termination")
				}
			})
		})
	}
}

func compactionItem(thread, turn, id string) map[string]any {
	return map[string]any{"threadId": thread, "turnId": turn, "item": map[string]any{"id": id, "type": "contextCompaction"}}
}
