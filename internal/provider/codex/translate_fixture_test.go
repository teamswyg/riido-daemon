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
