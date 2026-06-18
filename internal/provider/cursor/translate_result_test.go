package cursor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateResultSuccess(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"result","subtype":"success","result":"done","usage":{"input_tokens":1,"output_tokens":2}}`))
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}
}

func TestTranslateStepFinishUsageFallback(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"step_finish","usage":{"input_tokens":3,"output_tokens":4}}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventUsageDelta {
		t.Fatalf("usage fallback: %+v", evs)
	}
}

func TestTranslateMalformedWarning(t *testing.T) {
	evs := tx(t, agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: "malformed", Bytes: []byte("x")})
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventWarning {
		t.Fatalf("malformed: %+v", evs)
	}
}
