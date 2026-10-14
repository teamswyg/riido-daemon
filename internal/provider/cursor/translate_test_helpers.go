package cursor

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func tx(t *testing.T, raw agentbridge.RawEvent) []agentbridge.Event {
	t.Helper()
	evs, _, err := Translate(raw)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	return evs
}

func rawJSON(t *testing.T, s string) agentbridge.RawEvent {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	typ, _ := m["type"].(string)
	return agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: typ, Payload: m}
}
