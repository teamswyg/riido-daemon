package openclaw

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateFullResultError(t *testing.T) {
	raw := rawFull(t, `{"error":"model rejected"}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]

	if last.Kind != agentbridge.EventResult ||
		last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("err: %+v", last)
	}
}

func TestTranslateFullResultWithoutTextFailsClosed(t *testing.T) {
	raw := rawFull(t, `{
		"payloads":[],
		"meta":{
			"agentMeta":{
				"sessionId":"integration-openclaw",
				"usage":{"input":14886,"output":0,"total":14886}
			},
			"aborted":false
		}
	}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]

	if last.Kind != agentbridge.EventResult ||
		last.Result.Status != agentbridge.ResultFailed ||
		last.Result.Error == "" {
		t.Fatalf("empty full_result must fail closed: %+v", last)
	}
}
