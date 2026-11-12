package codex

import (
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

const waitIdle = 100 * time.Millisecond

func TestOwnedSessionContinuesWhileWaitedChildMakesProgress(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sess, child := startOwnedSessionWithIdle(t, waitIdle)
		emit := emitOwnedWaitNotification(child)
		emit("item/started", waitItemParams("parent-turn", "child"))
		for range 4 {
			time.Sleep(60 * time.Millisecond)
			select {
			case result := <-sess.Result():
				t.Fatalf("parent stopped while its requested review was progressing: status=%s error=%q", result.Status, result.Error)
			default:
			}
			emit("item/agentMessage/delta", waitChildDelta("child", "child-turn"))
		}
		time.Sleep(60 * time.Millisecond)
		emit("item/completed", map[string]any{
			"threadId": "child", "turnId": "child-turn",
			"item": map[string]any{"id": "command", "type": "commandExecution", "aggregatedOutput": "child command output"},
		})
		time.Sleep(60 * time.Millisecond)
		emit("turn/completed", map[string]any{
			"threadId": "child", "turn": map[string]any{"id": "child-turn", "status": "completed"},
		})
		time.Sleep(60 * time.Millisecond)
		emit("item/completed", waitItemParams("parent-turn", "child"))
		emit("item/completed", map[string]any{
			"threadId": "parent", "turnId": "parent-turn",
			"item": map[string]any{"id": "final", "type": "agentMessage", "phase": "final_answer", "text": "parent review"},
		})
		emit("turn/completed", map[string]any{
			"threadId": "parent", "turn": map[string]any{"id": "parent-turn", "status": "completed"},
		})
		if result := <-sess.Result(); result.Status != agentbridge.ResultCompleted || result.Output != "parent review" {
			t.Fatalf("parent did not finish its own review: %+v", result)
		}
		var logs []string
		for event := range sess.Events() {
			if event.Kind == agentbridge.EventSemanticActivity || strings.Contains(event.Text, "review finding") ||
				strings.Contains(event.Text, "child command output") {
				t.Fatalf("private child progress was published: %+v", event)
			}
			if event.Kind == agentbridge.EventLog {
				logs = append(logs, event.Text)
			}
		}
		diagnostics := strings.Join(logs, "\n")
		for _, want := range []string{"codex wait started", `receiver_threads=["child"]`, "reason=waited_target_terminal", "codex wait completed"} {
			if !strings.Contains(diagnostics, want) {
				t.Errorf("missing wait diagnostic %q in %s", want, diagnostics)
			}
		}
	})
}

func TestOwnedSessionIdleIgnoresUncorrelatedWaitActivity(t *testing.T) {
	for _, tc := range []struct {
		name        string
		waitTurn    string
		receiver    string
		setupMethod string
		childMethod string
		childTurn   string
	}{
		{name: "unrelated thread", waitTurn: "parent-turn", receiver: "other", childMethod: "item/agentMessage/delta", childTurn: "child-turn"},
		{name: "stale parent wait", waitTurn: "older-parent-turn", receiver: "child", childMethod: "item/agentMessage/delta", childTurn: "child-turn"},
		{name: "unidentified child turn", waitTurn: "parent-turn", receiver: "child", childMethod: "item/agentMessage/delta"},
		{name: "different child turn", waitTurn: "parent-turn", receiver: "child", setupMethod: "turn/started", childMethod: "item/agentMessage/delta", childTurn: "older-child-turn"},
		{name: "finished wait", waitTurn: "parent-turn", receiver: "child", setupMethod: "item/completed", childMethod: "item/agentMessage/delta", childTurn: "child-turn"},
		{name: "status without progress", waitTurn: "parent-turn", receiver: "child", childMethod: "thread/status/changed", childTurn: "child-turn"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				sess, child := startOwnedSessionWithIdle(t, waitIdle)
				emit := emitOwnedWaitNotification(child)
				emit("item/started", waitItemParams(tc.waitTurn, tc.receiver))
				switch tc.setupMethod {
				case "item/completed":
					emit(tc.setupMethod, waitItemParams(tc.waitTurn, tc.receiver))
				case "turn/started":
					emit(tc.setupMethod, map[string]any{
						"threadId": "child", "turn": map[string]any{"id": "child-turn", "status": "inProgress"},
					})
				}
				time.Sleep(60 * time.Millisecond)
				emit(tc.childMethod, waitChildDelta("child", tc.childTurn))
				time.Sleep(45 * time.Millisecond)
				select {
				case result := <-sess.Result():
					if result.Status != agentbridge.ResultTimeout || result.Error != "semantic idle timeout" {
						t.Fatalf("result=%+v", result)
					}
				default:
					t.Fatal("uncorrelated activity extended the parent's idle deadline")
				}
			})
		})
	}
}

func TestOwnedSessionHardTimeoutDuringActiveWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sess, child := startOwnedSessionWithIdle(t, waitIdle)
		emit := emitOwnedWaitNotification(child)
		emit("item/started", waitItemParams("parent-turn", "child"))
		for elapsed := time.Duration(0); elapsed < 5*time.Second; elapsed += 60 * time.Millisecond {
			time.Sleep(60 * time.Millisecond)
			select {
			case result := <-sess.Result():
				if result.Status != agentbridge.ResultTimeout || !strings.Contains(result.Error, "hard timeout") {
					t.Fatalf("result=%+v", result)
				}
				return
			default:
				emit("item/agentMessage/delta", waitChildDelta("child", "child-turn"))
			}
		}
		t.Fatal("child progress extended the parent's hard deadline")
	})
}

func emitOwnedWaitNotification(proc *process.FakeRunning) func(string, map[string]any) {
	return func(method string, params map[string]any) {
		raw := makeNotification(method, params)
		proc.EmitStdout(append(raw.Bytes, '\n'))
		synctest.Wait()
	}
}

func waitItemParams(parentTurn, receiver string) map[string]any {
	return map[string]any{
		"threadId": "parent", "turnId": parentTurn,
		"item": map[string]any{
			"id": "wait-item", "type": "collabAgentToolCall", "tool": "wait",
			"senderThreadId": "parent", "receiverThreadIds": []string{receiver}, "agentsStates": map[string]any{},
		},
	}
}

func waitChildDelta(thread, turn string) map[string]any {
	return map[string]any{"threadId": thread, "turnId": turn, "itemId": "review-output", "delta": "review finding"}
}
