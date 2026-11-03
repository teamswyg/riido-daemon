package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexParentRunIgnoresForeignEvents(t *testing.T) {
	for _, ids := range []struct{ name, thread, turn string }{
		{"child thread", "child", "child-turn"},
		{"other parent turn", "parent", "other-turn"},
	} {
		for _, method := range []string{
			"turn/started", "turn/completed", "turn/error", "error",
			"item/agentMessage/delta", "item/completed", "thread/status/changed",
			"thread/started", "thread/tokenUsage/updated",
		} {
			t.Run(ids.name+"/"+method, func(t *testing.T) {
				d, io, _ := parentResultFixture(t)
				events, commands, err := d.OnRaw(context.Background(), makeNotification(method, map[string]any{
					"threadId": ids.thread, "turnId": ids.turn, "status": "completed", "delta": "child text",
					"message": "child failed", "turn": map[string]any{"id": ids.turn, "status": "completed"},
					"item": map[string]any{"type": "agentMessage", "phase": "final_answer", "text": "child answer"},
				}), io)
				if err != nil || len(events) != 0 || len(commands) != 0 {
					t.Fatalf("foreign event escaped: events=%+v commands=%+v err=%v", events, commands, err)
				}
				if d.turnID != "parent-turn" || d.finalOutput != "" || d.lastRuntimeError != "" || d.sawAssistantOutput {
					t.Fatalf("foreign event mutated parent: %+v", d)
				}
			})
		}
	}
}

func TestCodexParentRunIgnoresForeignApprovalAndLegacyIdentity(t *testing.T) {
	for _, raw := range []agentbridge.RawEvent{
		makeServerRequest(42, "item/commandExecution/requestApproval", map[string]any{"threadId": "child", "turnId": "child-turn", "itemId": "cmd"}),
		makeNotification("turn_completed", map[string]any{"thread_id": "child", "turn_id": "child-turn", "output": "child answer"}),
		makeNotification("thread/started", map[string]any{"thread": map[string]any{"id": "child"}}),
	} {
		d, io, _ := parentResultFixture(t)
		events, _, err := d.OnRaw(context.Background(), raw, io)
		if err != nil || len(events) != 0 {
			t.Fatalf("foreign frame escaped: %+v err=%v", events, err)
		}
	}
}
