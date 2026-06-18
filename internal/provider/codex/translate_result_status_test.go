package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateTurnCompleted(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn_completed","params":{"output":"done"}}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("turn_completed: %+v", last)
	}
}

func TestTranslateTurnError(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn_error","params":{"message":"boom"}}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("turn_error: %+v", last)
	}
	if last.Result.Error != "boom" {
		t.Fatalf("error: %q", last.Result.Error)
	}
}
