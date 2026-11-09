package codex

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestCodexSessionWaitsForParentFinalResult(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	driver, err := NewProtocolDriver(agentbridge.StartRequest{})
	if err != nil {
		t.Fatal(err)
	}
	sess, err := session.Start(ctx, session.Config{
		TaskID: "task", RuntimeID: "codex", Adapter: integrationAdapter{},
		Process: fake, Spawn: process.Command{Executable: "fake-codex"}, ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	handshakeParentSession(t, ctx, running)
	for _, raw := range []agentbridge.RawEvent{
		makeNotification("item/agentMessage/delta", map[string]any{"threadId": "parent", "turnId": "parent-turn", "delta": "검토 중입니다"}),
		makeNotification("item/agentMessage/delta", map[string]any{"threadId": "child", "turnId": "child-turn", "delta": "지적 없음"}),
		makeNotification("turn/completed", map[string]any{"threadId": "child", "turn": map[string]any{"id": "child-turn", "status": "completed"}}),
		makeNotification("reasoning", map[string]any{"threadId": "parent", "turnId": "parent-turn", "text": "parent is still active"}),
	} {
		running.EmitStdout(append(raw.Bytes, '\n'))
	}
	waitForParentAfterChild(t, ctx, sess)
	for _, raw := range []agentbridge.RawEvent{parentMessage("final_answer", "전체 검토 결과입니다"), parentCompletion("completed")} {
		running.EmitStdout(append(raw.Bytes, '\n'))
	}
	select {
	case result := <-sess.Result():
		if result.Status != agentbridge.ResultCompleted || result.SessionID != "parent" || result.Output != "전체 검토 결과입니다" {
			t.Fatalf("wrong session result: %+v", result)
		}
	case <-ctx.Done():
		t.Fatal("parent final result was not delivered")
	}
}
