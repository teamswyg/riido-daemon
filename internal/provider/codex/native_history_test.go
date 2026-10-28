package codex

import (
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func nativeHistoryTurn(id, body, answer string) map[string]any {
	return map[string]any{"id": id, "status": "completed", "items": []any{
		map[string]any{"id": "user", "type": "userMessage", "content": []any{map[string]any{"type": "text", "text": body}}},
		map[string]any{"id": "tool", "type": "commandExecution", "text": "secret tool log"},
		map[string]any{"id": "answer", "type": "agentMessage", "text": answer},
	}}
}

func TestCodexResumeImportsCompleteHistoryBeforeNewTurn(t *testing.T) {
	for _, paginated := range []bool{false, true} {
		t.Run(map[bool]string{false: "inline", true: "paginated"}[paginated], func(t *testing.T) {
			req := agentbridge.StartRequest{ResumeSessionID: "native", Prompt: "# Riido AI Agent Assignment\n- task_id: task\nNew instruction", Metadata: map[string]string{"task_id": "task"}}
			d, err := NewProtocolDriver(req)
			if err != nil {
				t.Fatal(err)
			}
			io := newRecordingIO()
			if err := d.OnStart(t.Context(), io); err != nil {
				t.Fatal(err)
			}
			_ = io.next(t, time.Second)
			_, _, _ = d.OnRaw(t.Context(), makeResponse(1, nil), io)
			_ = io.next(t, time.Second)
			if frame := string(io.next(t, time.Second)); !strings.Contains(frame, `"method":"thread/resume"`) {
				t.Fatal(frame)
			}
			owned := nativeHistoryTurn("riido", "# Riido AI Agent Assignment\n- task_id: task\n", "existing")
			external := nativeHistoryTurn("external", "outside question", "outside answer")
			result := map[string]any{"thread": map[string]any{"id": "native", "turns": []any{owned, external}}}
			if paginated {
				result = map[string]any{"thread": map[string]any{"id": "native"}}
			}
			events, _, err := d.OnRaw(t.Context(), makeResponse(2, result), io)
			if err != nil {
				t.Fatal(err)
			}
			if paginated {
				if frame := string(io.next(t, time.Second)); !strings.Contains(frame, `"method":"thread/turns/list"`) || !strings.Contains(frame, `"itemsView":"full"`) {
					t.Fatal(frame)
				}
				events, _, err = d.OnRaw(t.Context(), makeResponse(3, map[string]any{"data": []any{owned}, "nextCursor": "page2"}), io)
				if err != nil || len(events) != 0 {
					t.Fatal("started before complete history", err)
				}
				if frame := string(io.next(t, time.Second)); !strings.Contains(frame, `"cursor":"page2"`) {
					t.Fatal(frame)
				}
				events, _, err = d.OnRaw(t.Context(), makeResponse(4, map[string]any{"data": []any{external}}), io)
			}
			if err != nil || len(events) != 3 || events[0].SessionID != "native" || events[1].NativeHistory.Body != "outside question" || events[2].NativeHistory.MessageID != "external:answer" {
				t.Fatalf("import: %+v %v", events, err)
			}
			if frame := string(io.next(t, time.Second)); !strings.Contains(frame, `"method":"turn/start"`) || !strings.Contains(frame, `"threadId":"native"`) {
				t.Fatal(frame)
			}
		})
	}
}

func TestCodexRejectsIncompleteOrRunningNativeTurns(t *testing.T) {
	for _, turn := range []any{map[string]any{"id": "missing-items"}, map[string]any{"id": "active", "status": "inProgress", "items": []any{}}} {
		if _, err := codexHistoryMessages("native", []any{turn}); err == nil {
			t.Fatal("incomplete turn accepted")
		}
	}
}

func TestCodexHistoryRetainsImageOnlyUserTurn(t *testing.T) {
	turn := nativeHistoryTurn("image", "", "image answer")
	turn["items"].([]any)[0].(map[string]any)["content"] = []any{map[string]any{"type": "image", "url": "data:private"}}
	messages, err := codexHistoryMessages("native", []any{turn})
	if err != nil || len(messages) != 2 || messages[0].Role != "user" || !strings.Contains(messages[0].Body, "attachment not imported") || strings.Contains(messages[0].Body, "private") {
		t.Fatalf("image-only turn lost or attachment exposed: %+v %v", messages, err)
	}
}
