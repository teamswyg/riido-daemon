package supervisor

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func readRunEvents(t *testing.T, path string) []ir.CanonicalEvent {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read run event log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	out := make([]ir.CanonicalEvent, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev ir.CanonicalEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("decode run event %q: %v", line, err)
		}
		out = append(out, ev)
	}
	return out
}

func assertRunEvent(t *testing.T, events []ir.CanonicalEvent, eventType ir.EventType, check func(ir.CanonicalEvent)) {
	t.Helper()
	for _, ev := range events {
		if ev.Type == eventType {
			if check != nil {
				check(ev)
			}
			return
		}
	}
	t.Fatalf("run event %s not found in %+v", eventType, events)
}

func TestProviderEventDraftMapsCatCEvents(t *testing.T) {
	for _, tc := range []struct {
		name string
		ev   agentbridge.Event
		want ir.EventType
	}{
		{"session", agentbridge.Event{Kind: agentbridge.EventSessionIdentified, SessionID: "s-1"}, ir.EventSessionPinned},
		{"text", agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hello"}, ir.EventTextDelta},
		{"thinking", agentbridge.Event{Kind: agentbridge.EventThinkingDelta, Text: "why"}, ir.EventReasoningDelta},
		{"tool-start", agentbridge.Event{Kind: agentbridge.EventToolCallStarted, Tool: agentbridge.ToolRef{ID: "tool-1", Name: "bash"}}, ir.EventToolCallStarted},
		{"tool-done", agentbridge.Event{Kind: agentbridge.EventToolCallCompleted, Tool: agentbridge.ToolRef{ID: "tool-1", Name: "bash"}}, ir.EventToolCallFinished},
		{"approval", agentbridge.Event{Kind: agentbridge.EventToolApprovalNeeded, Tool: agentbridge.ToolRef{ID: "approval-1", Kind: "exec"}}, ir.EventApprovalRequested},
		{"usage", agentbridge.Event{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{PromptTokens: 1}}, ir.EventUsageDelta},
		{"warning", agentbridge.Event{Kind: agentbridge.EventWarning, Text: "careful"}, ir.EventLogLine},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, payload, ok := providerEventDraft(tc.ev)
			if !ok {
				t.Fatalf("expected mapping for %+v", tc.ev)
			}
			if got != tc.want {
				t.Fatalf("event type = %s, want %s", got, tc.want)
			}
			if len(payload) == 0 {
				t.Fatalf("payload must not be empty")
			}
		})
	}
	if _, _, ok := providerEventDraft(agentbridge.Event{Kind: agentbridge.EventResult}); ok {
		t.Fatal("EventResult must stay outside non-transition Cat C mapping")
	}
}

func TestProviderEventDraftIncludesToolArgs(t *testing.T) {
	_, payload, ok := providerEventDraft(agentbridge.Event{
		Kind: agentbridge.EventToolCallStarted,
		Tool: agentbridge.ToolRef{
			ID:   "tool-1",
			Name: "Bash",
			Kind: "shell",
			Args: map[string]string{"command": "go test ./..."},
		},
	})
	if !ok {
		t.Fatal("expected tool event mapping")
	}
	args, ok := payload["args"].(map[string]string)
	if !ok {
		t.Fatalf("args payload type = %T", payload["args"])
	}
	if args["command"] != "go test ./..." {
		t.Fatalf("args payload = %+v", args)
	}
}

func TestTerminalResultDraftMapsTaskTransitions(t *testing.T) {
	for _, tc := range []struct {
		name string
		res  agentbridge.Result
		want ir.EventType
	}{
		{"completed", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}, ir.EventRunReportedDone},
		{"failed", agentbridge.Result{Status: agentbridge.ResultFailed, Error: "boom"}, ir.EventTaskFailed},
		{"blocked", agentbridge.Result{Status: agentbridge.ResultBlocked, Error: "capability"}, ir.EventTaskFailed},
		{"aborted", agentbridge.Result{Status: agentbridge.ResultAborted, Error: "exit"}, ir.EventTaskFailed},
		{"cancelled", agentbridge.Result{Status: agentbridge.ResultCancelled, Error: "user"}, ir.EventTaskCancelled},
		{"timeout", agentbridge.Result{Status: agentbridge.ResultTimeout, Error: "semantic idle timeout"}, ir.EventTaskTimedOut},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, payload := terminalResultDraft(tc.res)
			if got != tc.want {
				t.Fatalf("event type = %s, want %s", got, tc.want)
			}
			if len(payload) == 0 {
				t.Fatalf("payload must not be empty")
			}
			if !got.IsTransition() {
				t.Fatalf("%s must be an IR transition event", got)
			}
		})
	}
}
