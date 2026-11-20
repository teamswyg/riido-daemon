package codex

import (
	"bytes"
	"encoding/json"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestOwnedSessionRecoversOnlyCompletedParentAnswerOnIdle(t *testing.T) {
	for _, tc := range []struct {
		name, thread, turn, status, phase string
		want                              agentbridge.ResultStatus
	}{
		{"completed parent", "parent", "parent-turn", "completed", "final_answer", agentbridge.ResultCompleted},
		{"different thread", "child", "parent-turn", "completed", "final_answer", agentbridge.ResultTimeout},
		{"previous turn", "parent", "previous-turn", "completed", "final_answer", agentbridge.ResultTimeout},
		{"still running", "parent", "parent-turn", "inProgress", "final_answer", agentbridge.ResultTimeout},
		{"commentary only", "parent", "parent-turn", "completed", "commentary", agentbridge.ResultTimeout},
		{"missing completion timestamp", "parent", "parent-turn", "completed", "final_answer", agentbridge.ResultTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sess, child := startOwnedSessionWithObserver(t, 100*time.Millisecond, 2)
			// The native turn has saved a result, but the stream never delivered the
			// parent's completion. A child's reply must not complete this run.
			raw := makeNotification("item/agentMessage/delta", waitChildDelta("child", "child-turn"))
			child.EmitStdout(bytes.Repeat(append(raw.Bytes, '\n'), 300))
			select {
			case wire := <-child.StdinRecv():
				var request struct {
					ID     int64  `json:"id"`
					Method string `json:"method"`
					Params struct {
						ThreadID     string `json:"threadId"`
						IncludeTurns bool   `json:"includeTurns"`
					} `json:"params"`
				}
				if err := json.Unmarshal(wire, &request); err != nil || request.Method != "thread/read" || request.Params.ThreadID != "parent" || !request.Params.IncludeTurns {
					t.Fatalf("unexpected recovery request: %s (%v)", wire, err)
				}
				turn := map[string]any{
					"id": tc.turn, "status": tc.status, "completedAt": 1794468965,
					"items": []any{map[string]any{"id": "answer", "type": "agentMessage", "phase": tc.phase, "text": "parent review"}},
				}
				if tc.name == "missing completion timestamp" {
					delete(turn, "completedAt")
				}
				child.EmitStdout(jsonline(t, map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"thread": map[string]any{"id": tc.thread, "turns": []any{turn}}}}))
			case result := <-sess.Result():
				t.Fatalf("run ended without checking the native turn: %+v", result)
			case <-time.After(time.Second):
				t.Fatal("no idle recovery request")
			}
			result := <-sess.Result()
			if result.Status != tc.want || (tc.want == agentbridge.ResultCompleted && result.Output != "parent review") || (tc.want == agentbridge.ResultTimeout && result.Output != "") {
				t.Fatalf("result = %+v, want %s", result, tc.want)
			}
			if _, ok := <-sess.Result(); ok {
				t.Fatal("published more than one terminal result")
			}
		})
	}
}

func TestOwnedTerminalSurvivesDiagnosticBurstWithFullObserver(t *testing.T) {
	sess, child := startOwnedSessionWithObserver(t, time.Second, 2)
	diagnostic := makeNotification("item/agentMessage/delta", waitChildDelta("child", "child-turn"))
	child.EmitStdout(bytes.Repeat(append(diagnostic.Bytes, '\n'), 300))
	terminal := makeNotification("turn/completed", map[string]any{
		"threadId": "parent", "turn": map[string]any{
			"id": "parent-turn", "status": "completed",
			"items": []any{map[string]any{"id": "answer", "type": "agentMessage", "phase": "final_answer", "text": "parent review"}},
		},
	})
	child.EmitStdout(append(terminal.Bytes, '\n'))
	select {
	case result := <-sess.Result():
		if result.Status != agentbridge.ResultCompleted || result.Output != "parent review" {
			t.Fatalf("result = %+v", result)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("parent completion stalled behind diagnostic delivery")
	}
}

func TestOwnedIdleRecoveryIsBoundedAndCancellationPreemptsIt(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		name := "provider does not reply"
		if cancel {
			name = "user cancels during read"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				sess, child := startOwnedSessionWithIdle(t, 100*time.Millisecond)
				select {
				case <-child.StdinRecv():
				case <-time.After(time.Second):
					t.Fatal("no recovery read")
				}
				want := agentbridge.ResultTimeout
				if cancel {
					sess.Cancel(nil)
					want = agentbridge.ResultCancelled
				}
				select {
				case result := <-sess.Result():
					if result.Status != want {
						t.Fatalf("result = %+v, want %s", result, want)
					}
				case <-time.After(2500 * time.Millisecond):
					t.Fatal("idle recovery held the run open")
				}
			})
		})
	}
}
