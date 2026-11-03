package codex

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func handshakeParentSession(t *testing.T, ctx context.Context, running *process.FakeRunning) {
	t.Helper()
	for _, method := range []string{"initialize", "initialized", "thread/start", "turn/start"} {
		select {
		case body := <-running.StdinRecv():
			var frame map[string]any
			if err := json.Unmarshal(body, &frame); err != nil || frame["method"] != method {
				t.Fatalf("unexpected handshake: %s err=%v", body, err)
			}
			if method == "initialized" {
				continue
			}
			result := map[string]any{}
			if method == "thread/start" {
				result["thread"] = map[string]any{"id": "parent"}
			}
			if method == "turn/start" {
				result["turn"] = map[string]any{"id": "parent-turn", "status": "inProgress"}
			}
			id, _ := rpcID(frame)
			running.EmitStdout(append(makeResponse(id, result).Bytes, '\n'))
		case <-ctx.Done():
			t.Fatalf("handshake did not reach %s", method)
		}
	}
}

func waitForParentAfterChild(t *testing.T, ctx context.Context, sess *session.Session) {
	t.Helper()
	for {
		select {
		case event, ok := <-sess.Events():
			if !ok || event.Kind == agentbridge.EventResult || (event.Kind == agentbridge.EventTextDelta && event.Text == "지적 없음") {
				t.Fatalf("child escaped into parent session: %+v", event)
			}
			if event.Kind == agentbridge.EventThinkingDelta && event.Text == "parent is still active" {
				return
			}
		case result := <-sess.Result():
			t.Fatalf("child prematurely ended parent: %+v", result)
		case <-ctx.Done():
			t.Fatal("parent stopped processing after child completion")
		}
	}
}
