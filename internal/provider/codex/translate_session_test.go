package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateThreadStarted(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"thread_started","params":{"thread_id":"th-1"}}`)
	evs := tx(t, raw)
	if len(evs) < 1 || evs[0].Kind != agentbridge.EventSessionIdentified || evs[0].SessionID != "th-1" {
		t.Fatalf("thread_started: %+v", evs)
	}
}

func TestTranslateThreadStartedCurrentCodex(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"thread/started","params":{"thread":{"id":"th-1"}}}`)
	evs := tx(t, raw)
	if len(evs) < 1 || evs[0].Kind != agentbridge.EventSessionIdentified || evs[0].SessionID != "th-1" {
		t.Fatalf("thread/started: %+v", evs)
	}
}
