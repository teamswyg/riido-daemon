package codex

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func rawFromJSON(t *testing.T, s string) agentbridge.RawEvent {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("fixture parse %q: %v", s, err)
	}
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    classifyJSONRPC(m),
		Payload: m,
		Bytes:   []byte(s),
	}
}

func tx(t *testing.T, raw agentbridge.RawEvent) []agentbridge.Event {
	t.Helper()
	events, _, err := Translate(raw)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	return events
}

func TestTranslateThreadStarted(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"thread_started","params":{"thread_id":"th-1"}}`)
	evs := tx(t, raw)
	if len(evs) < 1 || evs[0].Kind != agentbridge.EventSessionIdentified || evs[0].SessionID != "th-1" {
		t.Fatalf("thread_started: %+v", evs)
	}
}

func TestTranslateThreadStartedCurrentCodex(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"thread/started","params":{"thread":{"id":"th-1"}}}`)
	evs := tx(t, raw)
	if len(evs) < 1 || evs[0].Kind != agentbridge.EventSessionIdentified || evs[0].SessionID != "th-1" {
		t.Fatalf("thread/started: %+v", evs)
	}
}

func TestTranslateTurnStarted(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn_started","params":{}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventLifecycle || evs[0].Phase != agentbridge.StateRunning {
		t.Fatalf("turn_started: %+v", evs)
	}
}

func TestTranslateAgentMessage(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"agent_message","params":{"text":"hello"}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventTextDelta || evs[0].Text != "hello" {
		t.Fatalf("agent_message: %+v", evs)
	}
}

func TestTranslateAgentMessageDeltaCurrentCodex(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"item/agentMessage/delta","params":{"delta":"hello"}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventTextDelta || evs[0].Text != "hello" {
		t.Fatalf("item/agentMessage/delta: %+v", evs)
	}
}

func TestTranslateReasoning(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"reasoning","params":{"text":"thinking..."}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventThinkingDelta {
		t.Fatalf("reasoning: %+v", evs)
	}
}

func TestTranslateCommandExecutionLifecycle(t *testing.T) {
	start := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"command_execution_started","params":{"id":"c1","command":"ls"}}`))
	if len(start) != 1 || start[0].Kind != agentbridge.EventToolCallStarted {
		t.Fatalf("start: %+v", start)
	}
	if start[0].Tool.ID != "c1" || start[0].Tool.Kind != "shell" {
		t.Fatalf("tool ref: %+v", start[0].Tool)
	}
	if start[0].Tool.Args["command"] != "ls" {
		t.Fatalf("tool args: %+v", start[0].Tool.Args)
	}

	delta := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"command_execution_output","params":{"id":"c1","chunk":"line1"}}`))
	if len(delta) != 1 || delta[0].Kind != agentbridge.EventToolCallDelta {
		t.Fatalf("delta: %+v", delta)
	}

	done := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"command_execution_completed","params":{"id":"c1","exit_code":0}}`))
	if len(done) != 1 || done[0].Kind != agentbridge.EventToolCallCompleted {
		t.Fatalf("done: %+v", done)
	}

	fail := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"command_execution_completed","params":{"id":"c2","exit_code":2}}`))
	if len(fail) != 1 || fail[0].Kind != agentbridge.EventToolCallFailed {
		t.Fatalf("fail: %+v", fail)
	}
}

func TestTranslateApplyPatch(t *testing.T) {
	start := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"apply_patch_started","params":{"id":"p1"}}`))
	if len(start) != 1 || start[0].Kind != agentbridge.EventToolCallStarted || start[0].Tool.Kind != "patch_apply" {
		t.Fatalf("patch start: %+v", start)
	}
	done := tx(t, rawFromJSON(t, `{"jsonrpc":"2.0","method":"apply_patch_completed","params":{"id":"p1"}}`))
	if len(done) != 1 || done[0].Kind != agentbridge.EventToolCallCompleted {
		t.Fatalf("patch done: %+v", done)
	}
}

func TestTranslateApprovalRequest(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":7,"method":"approve_command","params":{"command":"rm -rf /","id":"cmd-7"}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("approval: %+v", evs)
	}
	if evs[0].Tool.Args["command"] != "rm -rf /" {
		t.Fatalf("approval args: %+v", evs[0].Tool.Args)
	}
}

func TestTranslatePatchApprovalCapturesPathArg(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":8,"method":"approve_patch","params":{"path":".git/config","id":"patch-8"}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventToolApprovalNeeded {
		t.Fatalf("approval: %+v", evs)
	}
	if evs[0].Tool.Args["path"] != ".git/config" {
		t.Fatalf("patch args: %+v", evs[0].Tool.Args)
	}
}

func TestTranslateTurnCompleted(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn_completed","params":{"output":"done"}}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("turn_completed: %+v", last)
	}
}

func TestTranslateTurnError(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn_error","params":{"message":"boom"}}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("turn_error: %+v", last)
	}
	if last.Result.Error != "boom" {
		t.Fatalf("error: %q", last.Result.Error)
	}
}

func TestTranslateTurnErrorUsesNestedErrorMessage(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn/failed","params":{"error":{"message":"nested boom"}}}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("turn/failed: %+v", last)
	}
	if last.Result.Error != "nested boom" {
		t.Fatalf("error: %q", last.Result.Error)
	}
}

func TestTranslateUsage(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"usage","params":{"input_tokens":10,"output_tokens":20}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventUsageDelta {
		t.Fatalf("usage: %+v", evs)
	}
	if evs[0].Usage.PromptTokens != 10 || evs[0].Usage.CompletionTokens != 20 {
		t.Fatalf("usage tokens: %+v", evs[0].Usage)
	}
}

func TestTranslateCurrentCodexUsage(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"thread/tokenUsage/updated","params":{"tokenUsage":{"total":{"inputTokens":10,"cachedInputTokens":3,"outputTokens":4,"reasoningOutputTokens":2}}}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventUsageDelta {
		t.Fatalf("usage event: %+v", evs)
	}
	if evs[0].Usage.PromptTokens != 10 || evs[0].Usage.CacheReadTokens != 3 || evs[0].Usage.CompletionTokens != 4 || evs[0].Usage.ReasoningTokens != 2 {
		t.Fatalf("usage: %+v", evs[0].Usage)
	}
}

func TestTranslateCodexInternalNotificationsAreDropped(t *testing.T) {
	for _, method := range []string{
		"account/rateLimits/updated",
		"account_rate_limits_updated",
		"item/started",
		"item/completed",
		"remoteControl/status/changed",
	} {
		t.Run(method, func(t *testing.T) {
			raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"`+method+`","params":{}}`)
			evs := tx(t, raw)
			if len(evs) != 0 {
				t.Fatalf("%s should not surface user-visible events: %+v", method, evs)
			}
		})
	}
}

func TestTranslateMalformedProducesWarning(t *testing.T) {
	raw := agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: "malformed", Bytes: []byte("junk")}
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventWarning {
		t.Fatalf("malformed: %+v", evs)
	}
}

func TestTranslateUnknownNotificationLogged(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"some_new_event","params":{}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventLog {
		t.Fatalf("unknown: %+v", evs)
	}
}

func TestTranslateErrorResponse(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"boom"}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventError {
		t.Fatalf("err response: %+v", evs)
	}
}
