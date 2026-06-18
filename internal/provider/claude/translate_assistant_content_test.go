package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateAssistantTextDelta(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"assistant","message":{"content":[{"type":"text","text":"hello"}]}}`)
	events := translate(t, raw)

	if len(events) != 1 ||
		events[0].Kind != agentbridge.EventTextDelta ||
		events[0].Text != "hello" {
		t.Fatalf("text delta: %+v", events)
	}
}

func TestTranslateAssistantThinkingDelta(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"reasoning..."}]}}`)
	events := translate(t, raw)

	if len(events) != 1 || events[0].Kind != agentbridge.EventThinkingDelta {
		t.Fatalf("thinking: %+v", events)
	}
}
