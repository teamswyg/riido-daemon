package codex

import (
	"encoding/json"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestOwnedIdleContinuesVerifiedRunningTurn(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sess, child := startOwnedSessionWithIdle(t, waitIdle)
		requestID := receiveIdleRead(t, child)
		child.EmitStdout(idleReadReply(t, requestID, "parent", "parent-turn", "inProgress"))
		synctest.Wait()
		time.Sleep(waitIdle / 2)
		select {
		case result := <-sess.Result():
			t.Fatalf("still-running owned turn was terminated before recovery: %+v", result)
		default:
		}
		emitIdleCompletion(t, child)
		if result := <-sess.Result(); result.Status != agentbridge.ResultCompleted || result.Output != "recovered parent answer" {
			t.Fatalf("result = %+v", result)
		}
	})
}

func TestOwnedIdlePreservesCompletionWhileReading(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sess, child := startOwnedSessionWithIdle(t, waitIdle)
		receiveIdleRead(t, child)
		// The stream recovers before the read response. Do not discard its answer.
		emitIdleCompletion(t, child)
		if result := <-sess.Result(); result.Status != agentbridge.ResultCompleted || result.Output != "recovered parent answer" {
			t.Fatalf("result = %+v", result)
		}
	})
}

func receiveIdleRead(t *testing.T, child *process.FakeRunning) any {
	t.Helper()
	select {
	case wire := <-child.StdinRecv():
		var request map[string]any
		if err := json.Unmarshal(wire, &request); err != nil || request["method"] != "thread/read" {
			t.Fatalf("unexpected idle request %s: %v", wire, err)
		}
		return request["id"]
	case <-time.After(10 * time.Second):
		t.Fatal("no idle read")
		return nil
	}
}

func idleReadReply(t *testing.T, id any, thread, turn, status string) []byte {
	t.Helper()
	return jsonline(t, map[string]any{"jsonrpc": "2.0", "id": id, "result": map[string]any{"thread": map[string]any{
		"id": thread, "turns": []any{map[string]any{"id": turn, "status": status}},
	}}})
}

func emitIdleCompletion(t *testing.T, child *process.FakeRunning) {
	t.Helper()
	emitOwnedWaitNotification(child)("turn/completed", map[string]any{
		"threadId": "parent", "turn": map[string]any{
			"id": "parent-turn", "status": "completed",
			"items": []any{map[string]any{"id": "answer", "type": "agentMessage", "phase": "final_answer", "text": "recovered parent answer"}},
		},
	})
}
