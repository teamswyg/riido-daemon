package codex

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestOwnedIdlePreservesSeparateAnswerWithFullObserver(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sess, child := startOwnedSessionWithObserver(t, waitIdle, 2)
		receiveIdleRead(t, child)
		var wire []byte
		add := func(method string, p map[string]any) {
			raw := makeNotification(method, p)
			wire = append(wire, raw.Bytes...)
			wire = append(wire, '\n')
		}
		for _, p := range []map[string]any{
			{"threadId": "child", "turn": map[string]any{"id": "child-turn", "status": "completed"}},
			{"threadId": "parent", "turn": map[string]any{"id": "old-turn", "status": "completed"}},
			{"threadId": "parent", "turn": map[string]any{"status": "completed"}},
			{"turn": map[string]any{"id": "parent-turn", "status": "completed"}},
		} {
			add("turn/completed", p)
		}
		add("item/completed", map[string]any{"threadId": "parent", "turnId": "parent-turn", "item": map[string]any{
			"id": "answer", "type": "agentMessage", "phase": "final_answer", "text": "separate answer",
		}})
		add("turn/completed", map[string]any{"threadId": "parent", "turn": map[string]any{"id": "parent-turn", "status": "completed"}})
		child.EmitStdout(wire)
		select {
		case result := <-sess.Result():
			if result.Status != agentbridge.ResultCompleted || result.Output != "separate answer" {
				t.Fatalf("result = %+v", result)
			}
		case <-time.After(time.Second):
			t.Fatal("completion blocked behind observer")
		}
	})
}

func TestOwnedIdleReadDeadlineSurvivesWarningBackpressure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sess, child := startOwnedSessionWithObserver(t, waitIdle, 2)
		receiveIdleRead(t, child)
		raw := makeNotification("error", map[string]any{"threadId": "parent", "turnId": "parent-turn", "willRetry": true, "error": map[string]any{"message": "temporary connection loss"}})
		child.EmitStdout(append(raw.Bytes, '\n'))
		select {
		case result := <-sess.Result():
			if result.Status != agentbridge.ResultTimeout || result.Error != "semantic idle timeout" {
				t.Fatalf("result = %+v", result)
			}
		case <-time.After(2100 * time.Millisecond):
			t.Fatal("warning delivery bypassed recovery read deadline")
		}
	})
}
