package codex

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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

// turn/failed is the newer codex name for a failed turn → terminal failure.
func TestTranslateTurnFailed(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn/failed","params":{"message":"nope"}}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed || last.Result.Error != "nope" {
		t.Fatalf("turn/failed: %+v", last)
	}
}

// New codex app-server lifecycle/rate-limit notifications are recognized as
// informational Logs — never "unknown notification", and crucially never a
// terminal EventResult (a per-item completion must not truncate a live turn).
func TestTranslateNewLifecycleNotificationsAreNonTerminalLogs(t *testing.T) {
	methods := []string{
		"item/started", "item/updated", "item/completed",
		"hook/started", "hook/completed",
		"mcpServer/startupStatus/updated",
		"remoteControl/status/changed",
		"account/rateLimits/updated",
	}
	for _, m := range methods {
		raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"`+m+`","params":{}}`)
		evs := tx(t, raw)
		if len(evs) != 1 || evs[0].Kind != agentbridge.EventLog {
			t.Fatalf("%s: want single EventLog, got %+v", m, evs)
		}
		if strings.Contains(evs[0].Text, "unknown") {
			t.Fatalf("%s: still labeled unknown: %q", m, evs[0].Text)
		}
	}
}
