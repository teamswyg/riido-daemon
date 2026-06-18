package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateTurnStarted(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn_started","params":{}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventLifecycle || evs[0].Phase != agentbridge.StateRunning {
		t.Fatalf("turn_started: %+v", evs)
	}
}
