package openclaw

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateFullResultSuccess(t *testing.T) {
	raw := rawFull(t, `{"session_id":"sess-1","text":"hello world","usage":{"prompt_tokens":3,"completion_tokens":7}}`)
	evs := tx(t, raw)

	if len(evs) == 0 {
		t.Fatal("no events")
	}
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult ||
		last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}

	sawSession, sawUsage, sawText := false, false, false
	for _, ev := range evs {
		switch ev.Kind {
		case agentbridge.EventSessionIdentified:
			if ev.SessionID == "sess-1" {
				sawSession = true
			}
		case agentbridge.EventUsageDelta:
			if ev.Usage.PromptTokens == 3 && ev.Usage.CompletionTokens == 7 {
				sawUsage = true
			}
		case agentbridge.EventTextDelta:
			if ev.Text == "hello world" {
				sawText = true
			}
		}
	}
	if !sawSession || !sawUsage || !sawText {
		t.Fatalf("missing session/usage/text in events: %+v", evs)
	}
}
