package cursor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateSystemInit(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"system","subtype":"init","session_id":"sess-1"}`))
	if len(evs) < 2 || evs[0].Kind != agentbridge.EventSessionIdentified {
		t.Fatalf("system: %+v", evs)
	}
}

func TestTranslateText(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"text","text":"hello"}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventTextDelta || evs[0].Text != "hello" {
		t.Fatalf("text: %+v", evs)
	}
}

func TestTranslateAssistantText(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"assistant","content":[{"type":"output_text","text":"x"}]}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventTextDelta || evs[0].Text != "x" {
		t.Fatalf("assistant text: %+v", evs)
	}
}

func TestTranslateAssistantThinking(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"assistant","content":[{"type":"thinking","text":"hmm"}]}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventThinkingDelta {
		t.Fatalf("thinking: %+v", evs)
	}
}
