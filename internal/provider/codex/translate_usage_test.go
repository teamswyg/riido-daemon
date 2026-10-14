package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
	if evs[0].Usage.PromptTokens != 10 || evs[0].Usage.CacheReadTokens != 3 ||
		evs[0].Usage.CompletionTokens != 4 || evs[0].Usage.ReasoningTokens != 2 {
		t.Fatalf("usage: %+v", evs[0].Usage)
	}
}
