package claude

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func mustParseRaw(t *testing.T, payload string) agentbridge.RawEvent {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatalf("parse fixture %q: %v", payload, err)
	}
	typ, _ := m["type"].(string)
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    typ,
		Payload: m,
		Bytes:   []byte(payload),
	}
}

func translate(t *testing.T, raw agentbridge.RawEvent) []agentbridge.Event {
	t.Helper()
	events, _, err := Translate(raw)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	return events
}

// system/init → SessionIdentified + Lifecycle(Running)
func TestTranslateSystemInitProducesSessionAndLifecycle(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"system","subtype":"init","session_id":"sess-42"}`)
	events := translate(t, raw)
	if len(events) < 2 {
		t.Fatalf("want >=2 events, got %d: %+v", len(events), events)
	}
	if events[0].Kind != agentbridge.EventSessionIdentified || events[0].SessionID != "sess-42" {
		t.Fatalf("first event: %+v", events[0])
	}
	if events[1].Kind != agentbridge.EventLifecycle || events[1].Phase != agentbridge.StateRunning {
		t.Fatalf("second event: %+v", events[1])
	}
}

// assistant content text → TextDelta
func TestTranslateAssistantTextDelta(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"assistant","message":{"content":[{"type":"text","text":"hello"}]}}`)
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta || events[0].Text != "hello" {
		t.Fatalf("text delta: %+v", events)
	}
}

// assistant content thinking → ThinkingDelta
func TestTranslateAssistantThinkingDelta(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"reasoning..."}]}}`)
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventThinkingDelta {
		t.Fatalf("thinking: %+v", events)
	}
}

// assistant tool_use → ToolCallStarted
func TestTranslateAssistantToolUse(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"tu_1","name":"Edit","input":{"file_path":".git/config","api_token":"raw","note":"ghp_`+strings.Repeat("a", 20)+`"}}]}}`)
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventToolCallStarted {
		t.Fatalf("tool_use: %+v", events)
	}
	if events[0].Tool.ID != "tu_1" || events[0].Tool.Name != "Edit" {
		t.Fatalf("tool ref: %+v", events[0].Tool)
	}
	if events[0].Tool.Args["file_path"] != ".git/config" {
		t.Fatalf("tool args: %+v", events[0].Tool.Args)
	}
	if events[0].Tool.Args["api_token"] != "[redacted]" {
		t.Fatalf("sensitive args must be redacted: %+v", events[0].Tool.Args)
	}
	if events[0].Tool.Args["note"] != "[redacted]" {
		t.Fatalf("secret-looking value must be redacted: %+v", events[0].Tool.Args)
	}
}

// user/tool_result → ToolCallCompleted (or ToolCallFailed on is_error=true)
func TestTranslateUserToolResultCompleted(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"tu_1"}]}}`)
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventToolCallCompleted {
		t.Fatalf("tool_result: %+v", events)
	}
	if events[0].Tool.ID != "tu_1" {
		t.Fatalf("tool id: %+v", events[0].Tool)
	}
}

func TestTranslateUserToolResultFailed(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"tu_1","is_error":true}]}}`)
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventToolCallFailed {
		t.Fatalf("tool_result err: %+v", events)
	}
}

// result → Result event with status (and usage if present)
func TestTranslateResultSuccess(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"result","subtype":"success","result":"done","usage":{"input_tokens":3,"output_tokens":7}}`)
	events := translate(t, raw)
	if len(events) == 0 {
		t.Fatalf("expected events")
	}
	// last event must be Result.
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}
	if last.Result.Output != "done" {
		t.Fatalf("output: %q", last.Result.Output)
	}
	// usage may be a separate UsageDelta event preceding Result.
	sawUsage := false
	for _, ev := range events {
		if ev.Kind == agentbridge.EventUsageDelta {
			sawUsage = true
			if ev.Usage.PromptTokens != 3 || ev.Usage.CompletionTokens != 7 {
				t.Fatalf("usage tokens: %+v", ev.Usage)
			}
		}
	}
	if !sawUsage {
		t.Fatalf("expected usage delta in events: %+v", events)
	}
}

func TestTranslateResultError(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"result","subtype":"error","error":"boom"}`)
	events := translate(t, raw)
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("result err: %+v", last)
	}
	if last.Result.Error != "boom" {
		t.Fatalf("error: %q", last.Result.Error)
	}
}

// control_request must NOT be silently dropped (spec §15 item 3).
func TestTranslateControlRequestProducesApprovalNeeded(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"control_request","request_id":"r_1","request":{"subtype":"permission_request","tool_use_id":"tu_1","tool_name":"Bash","tool_input":{"command":"terraform destroy"}}}`)
	events := translate(t, raw)
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d: %+v", len(events), events)
	}
	if events[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("control_request must produce approval event, got %s", events[0].Kind)
	}
	if events[0].Tool.ID != "tu_1" || events[0].Tool.Name != "Bash" {
		t.Fatalf("tool ref: %+v", events[0].Tool)
	}
	if events[0].Tool.ProviderRequestID != "r_1" {
		t.Fatalf("provider request id: %+v", events[0].Tool)
	}
	if events[0].Tool.Args["command"] != "terraform destroy" {
		t.Fatalf("tool args: %+v", events[0].Tool.Args)
	}
}
