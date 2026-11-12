package codex

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestOwnedSessionChildFinishesBeforeParent(t *testing.T) {
	for _, status := range []string{"completed", "failed", "interrupted"} {
		t.Run(status, func(t *testing.T) {
			sess, child := startOwnedSession(t)
			emit := func(raw agentbridge.RawEvent) { child.EmitStdout(append(raw.Bytes, '\n')) }
			for _, method := range []string{"turn/started", "item/agentMessage/delta", "error", "turn/completed"} {
				emit(makeNotification(method, map[string]any{"threadId": "child", "thread": map[string]any{"id": "child"}, "turnId": "child-turn", "turn": map[string]any{"id": "child-turn", "status": "completed"}, "delta": "child review", "error": map[string]any{"message": "child error"}}))
			}
			// Receiving this marker proves the session processed the preceding
			// child terminal event and let the parent's remaining work continue.
			emit(makeNotification("item/agentMessage/delta", map[string]any{"threadId": "parent", "turnId": "parent-turn", "itemId": "progress", "delta": "parent posting review"}))
			for waiting := true; waiting; {
				select {
				case result := <-sess.Result():
					t.Fatalf("child completed/killed parent before posting: %+v", result)
				case event, ok := <-sess.Events():
					if !ok {
						t.Fatal("parent event stream ended")
					}
					waiting = event.Kind != agentbridge.EventProgress || event.Text != "parent posting review"
				case <-time.After(2 * time.Second):
					t.Fatal("parent did not reach posting step")
				}
			}
			select {
			case <-child.KillRecv():
				t.Fatal("child completion killed parent process")
			default:
			}
			emit(makeNotification("item/completed", map[string]any{"threadId": "parent", "turnId": "parent-turn", "item": map[string]any{"id": "final", "type": "agentMessage", "phase": "final_answer", "text": "parent final review"}}))
			emit(makeNotification("turn/completed", map[string]any{"threadId": "parent", "turn": map[string]any{"id": "parent-turn", "status": status, "error": map[string]any{"message": "parent failure"}}}))
			var logs []string
			for event := range sess.Events() {
				if event.Kind == agentbridge.EventLog {
					logs = append(logs, event.Text)
				}
			}
			result := <-sess.Result()
			want := map[string]agentbridge.ResultStatus{"completed": agentbridge.ResultCompleted, "failed": agentbridge.ResultFailed, "interrupted": agentbridge.ResultCancelled}[status]
			if result.Status != want || result.Output != "parent final review" || result.SessionID != "parent" {
				t.Fatalf("result=%+v", result)
			}
			log := strings.Join(logs, "\n")
			for _, field := range []string{"decision=terminal", `target_thread="parent"`, `target_turn="parent-turn"`, `event="notification:turn/completed"`, "reason=" + string(want)} {
				if !strings.Contains(log, field) {
					t.Fatalf("missing diagnostic %s: %s", field, log)
				}
			}
		})
	}
}

func startOwnedSession(t *testing.T) (*session.Session, *process.FakeRunning) {
	t.Helper()
	return startOwnedSessionWithIdle(t, 0)
}

func startOwnedSessionWithIdle(t *testing.T, idle time.Duration) (*session.Session, *process.FakeRunning) {
	t.Helper()
	return startOwnedSessionWithObserver(t, idle, 0)
}

func startOwnedSessionWithObserver(t *testing.T, idle time.Duration, buffer int) (*session.Session, *process.FakeRunning) {
	t.Helper()
	child := process.NewFakeRunning()
	driver, _ := NewProtocolDriver(agentbridge.StartRequest{Prompt: "review"})
	sess, err := session.Start(t.Context(), session.Config{TaskID: "owned", RuntimeID: "codex", Adapter: integrationAdapter{}, Process: &process.Fake{NextRunning: child}, ProtocolDriver: driver, HardTimeout: 5 * time.Second, SemanticIdle: idle, EventBuffer: buffer})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sess.Cancel(nil); <-sess.Done() })
	for _, method := range []string{"initialize", "initialized", "thread/start", "turn/start"} {
		select {
		case wire := <-child.StdinRecv():
			var request map[string]any
			if err := json.Unmarshal(wire, &request); err != nil {
				t.Fatal(err)
			}
			if request["method"] != method {
				t.Fatalf("expected %s got %s", method, wire)
			}
			result := map[string]any{}
			if method == "thread/start" {
				result["thread"] = map[string]any{"id": "parent"}
			}
			if method == "turn/start" {
				result["turn"] = map[string]any{"id": "parent-turn", "status": "inProgress"}
			}
			if method != "initialized" {
				child.EmitStdout(jsonline(t, map[string]any{"jsonrpc": "2.0", "id": request["id"], "result": result}))
			}
		case <-time.After(time.Second):
			t.Fatalf("missing %s request", method)
		}
	}
	return sess, child
}
