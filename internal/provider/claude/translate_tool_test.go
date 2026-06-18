package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
