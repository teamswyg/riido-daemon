package openclaw

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateCurrentFullResultShape(t *testing.T) {
	raw := rawFull(t, `{
		"payloads":[{"text":"ok","mediaUrl":null}],
		"meta":{
			"agentMeta":{
				"sessionId":"integration-openclaw",
				"usage":{"input":14886,"output":2,"total":14888},
				"lastCallUsage":{"input":14886,"output":2,"cacheRead":0,"cacheWrite":0,"total":14888}
			},
			"aborted":false
		}
	}`)
	evs := tx(t, raw)

	var saw struct {
		session bool
		usage   bool
		text    bool
		result  bool
	}
	for _, ev := range evs {
		switch ev.Kind {
		case agentbridge.EventSessionIdentified:
			saw.session = ev.SessionID == "integration-openclaw"
		case agentbridge.EventUsageDelta:
			saw.usage = ev.Usage.PromptTokens == 14886 &&
				ev.Usage.CompletionTokens == 2
		case agentbridge.EventTextDelta:
			saw.text = ev.Text == "ok"
		case agentbridge.EventResult:
			saw.result = ev.Result.Status == agentbridge.ResultCompleted &&
				ev.Result.Output == "ok"
		}
	}
	if !saw.session || !saw.usage || !saw.text || !saw.result {
		t.Fatalf("current full_result shape coverage gap: %+v events=%+v", saw, evs)
	}
}
