package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateMalformedProducesWarning(t *testing.T) {
	raw := agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: "malformed", Bytes: []byte("junk")}
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventWarning {
		t.Fatalf("malformed: %+v", evs)
	}
}

func TestTranslateUnknownNotificationLogged(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"some_new_event","params":{}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventLog {
		t.Fatalf("unknown: %+v", evs)
	}
}

func TestTranslateErrorResponse(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"boom"}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventError {
		t.Fatalf("err response: %+v", evs)
	}
}
