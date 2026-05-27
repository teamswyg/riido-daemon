package openclaw

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// TestGoldenFullResultJSON feeds testdata/full_result.json through the
// real parser (single-chunk full-JSON mode) and asserts the translator
// surfaces SessionIdentified + UsageDelta + TextDelta + Result(completed).
//
// OpenClaw's full-result output is one big JSON object with no
// internal newlines. The parser switches into NDJSON mode the moment
// it sees a newline-terminated line — so the fixture, which carries a
// trailing newline as good text-file hygiene, is normalized here by
// stripping the trailing \n before feeding. A real codex/openclaw
// process never emits that trailing newline before EOF, so this matches
// production behavior.
func TestGoldenFullResultJSON(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "full_result.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	body = bytes.TrimRight(body, "\n")

	p := NewParser()
	feed, err := p.FeedStdout(body)
	if err != nil {
		t.Fatalf("FeedStdout: %v", err)
	}
	closed, _ := p.Close()
	raws := append([]agentbridge.RawEvent{}, feed...)
	raws = append(raws, closed...)

	var saw struct {
		session, usage, text, result bool
	}
	for _, raw := range raws {
		evs, _, _ := Translate(raw)
		for _, ev := range evs {
			switch ev.Kind {
			case agentbridge.EventSessionIdentified:
				saw.session = true
			case agentbridge.EventUsageDelta:
				saw.usage = true
			case agentbridge.EventTextDelta:
				saw.text = true
			case agentbridge.EventResult:
				if ev.Result.Status == agentbridge.ResultCompleted {
					saw.result = true
				}
			}
		}
	}
	if !(saw.session && saw.usage && saw.text && saw.result) {
		t.Fatalf("full_result coverage gap: %+v", saw)
	}
}

// TestGoldenNDJSONResultJSONL feeds testdata/ndjson_result.jsonl through
// the parser in NDJSON mode and asserts the streaming event sequence
// is translated.
func TestGoldenNDJSONResultJSONL(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "ndjson_result.jsonl"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	p := NewParser()
	feed, err := p.FeedStdout(body)
	if err != nil {
		t.Fatalf("FeedStdout: %v", err)
	}
	closed, _ := p.Close()
	raws := append([]agentbridge.RawEvent{}, feed...)
	raws = append(raws, closed...)

	var saw struct {
		session, text, log, usage bool
	}
	for _, raw := range raws {
		evs, _, _ := Translate(raw)
		for _, ev := range evs {
			switch ev.Kind {
			case agentbridge.EventSessionIdentified:
				saw.session = true
			case agentbridge.EventTextDelta:
				saw.text = true
			case agentbridge.EventLog:
				saw.log = true
			case agentbridge.EventUsageDelta:
				saw.usage = true
			}
		}
	}
	if !(saw.session && saw.text && saw.log && saw.usage) {
		t.Fatalf("ndjson_result coverage gap: %+v", saw)
	}
}
