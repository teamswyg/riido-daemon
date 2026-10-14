package claude

import (
	"encoding/json"
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
