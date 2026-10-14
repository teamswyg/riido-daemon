package openclaw

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateNDJSONText(t *testing.T) {
	raw := agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "ndjson:text",
		Payload: map[string]any{"event": "text", "text": "chunk"},
	}
	evs := tx(t, raw)

	if len(evs) != 1 ||
		evs[0].Kind != agentbridge.EventTextDelta ||
		evs[0].Text != "chunk" {
		t.Fatalf("ndjson text: %+v", evs)
	}
}

func TestTranslateMalformedWarning(t *testing.T) {
	raw := agentbridge.RawEvent{
		Source: agentbridge.RawSourceStdout,
		Type:   "malformed",
		Bytes:  []byte("x"),
	}
	evs := tx(t, raw)

	if len(evs) != 1 || evs[0].Kind != agentbridge.EventWarning {
		t.Fatalf("malformed: %+v", evs)
	}
}
