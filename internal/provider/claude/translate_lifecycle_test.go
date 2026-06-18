package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateSystemInitProducesSessionAndLifecycle(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"system","subtype":"init","session_id":"sess-42"}`)
	events := translate(t, raw)

	if len(events) < 2 {
		t.Fatalf("want >=2 events, got %d: %+v", len(events), events)
	}
	if events[0].Kind != agentbridge.EventSessionIdentified ||
		events[0].SessionID != "sess-42" {
		t.Fatalf("first event: %+v", events[0])
	}
	if events[1].Kind != agentbridge.EventLifecycle ||
		events[1].Phase != agentbridge.StateRunning {
		t.Fatalf("second event: %+v", events[1])
	}
}
