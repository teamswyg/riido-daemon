package openclaw

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestParserFeedStderrEmitsTrimmedLines(t *testing.T) {
	p := NewParser()
	raws, err := p.FeedStderr([]byte("  first warning"))
	if err != nil {
		t.Fatal(err)
	}
	if len(raws) != 0 {
		t.Fatalf("partial stderr emitted raw events: %+v", raws)
	}
	raws, err = p.FeedStderr([]byte("\n\n second warning \n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(raws) != 2 {
		t.Fatalf("raw count=%d events=%+v", len(raws), raws)
	}
	for i, want := range []string{"first warning", "second warning"} {
		if raws[i].Source != agentbridge.RawSourceStderr ||
			raws[i].Type != "stderr" ||
			string(raws[i].Bytes) != want {
			t.Fatalf("raw[%d]=%+v, want %q stderr", i, raws[i], want)
		}
	}
}

func TestTranslateUnknownNDJSONEventBecomesLog(t *testing.T) {
	raw := agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "ndjson:surprise",
		Payload: map[string]any{"message": "ignored"},
	}
	evs := tx(t, raw)
	if len(evs) != 1 ||
		evs[0].Kind != agentbridge.EventLog ||
		evs[0].Text != "openclaw ndjson unknown event: surprise" {
		t.Fatalf("unknown ndjson events=%+v", evs)
	}
}

func TestTranslateFullResultUsesLastCallUsageFallback(t *testing.T) {
	raw := rawFull(t, `{
		"text":"ok",
		"meta":{"agentMeta":{"lastCallUsage":{"input":5,"output":8,"cacheRead":2}}}
	}`)
	evs := tx(t, raw)
	for _, ev := range evs {
		if ev.Kind == agentbridge.EventUsageDelta &&
			ev.Usage.PromptTokens == 5 &&
			ev.Usage.CompletionTokens == 8 &&
			ev.Usage.CacheReadTokens == 2 {
			return
		}
	}
	t.Fatalf("lastCallUsage fallback not translated: %+v", evs)
}
