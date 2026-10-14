package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexRecoveryBoundary(t *testing.T) {
	cases := []struct {
		name          string
		beforeText    bool
		errorParams   map[string]any
		completion    map[string]any
		want          agentbridge.ResultStatus
		wantErrorKind agentbridge.EventKind
	}{
		{"native retry empty success", false, map[string]any{"error": map[string]any{"message": "transient disconnect"}, "willRetry": true}, map[string]any{"turn": map[string]any{"status": "completed"}}, agentbridge.ResultCompleted, agentbridge.EventWarning},
		{"interruption after error", false, map[string]any{"message": "disconnect"}, map[string]any{"turn": map[string]any{"status": "interrupted"}}, agentbridge.ResultCancelled, agentbridge.EventError},
		{"partial text before later failure", true, map[string]any{"message": "stream failed"}, map[string]any{"output": ""}, agentbridge.ResultFailed, agentbridge.EventError},
		{"native empty success after error", false, map[string]any{"message": "disconnect"}, map[string]any{"turn": map[string]any{"status": "completed"}}, agentbridge.ResultCompleted, agentbridge.EventError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, io := startedProtocolDriver(t)
			if c.beforeText {
				_, _, _ = d.OnRaw(context.Background(), makeNotification("agent_message", map[string]any{"text": "checking the task"}), io)
			}
			events, _, err := d.OnRaw(context.Background(), makeNotification("error", c.errorParams), io)
			if err != nil || len(events) != 1 || events[0].Kind != c.wantErrorKind {
				t.Fatalf("error events=%+v err=%v", events, err)
			}
			events, _, err = d.OnRaw(context.Background(), makeNotification("turn/completed", c.completion), io)
			if err != nil || len(events) == 0 {
				t.Fatalf("completion=%+v err=%v", events, err)
			}
			last := events[len(events)-1]
			if last.Kind != agentbridge.EventResult || last.Result.Status != c.want {
				t.Fatalf("completion=%+v want=%s", events, c.want)
			}
		})
	}
}
