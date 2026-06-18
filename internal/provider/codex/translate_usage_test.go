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
