package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
