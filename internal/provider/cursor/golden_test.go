package cursor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// runFixtureThroughParser feeds an entire fixture through Parser +
// Translate and returns every Event produced. We exercise the parser
// path (not Translate alone) so stream-prefix normalization is tested
// at the same time.
func runFixtureThroughParser(t *testing.T, name string) []agentbridge.Event {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	p := NewParser()
	raws, err := p.FeedStdout(body)
	if err != nil {
		t.Fatalf("FeedStdout: %v", err)
	}
	closed, _ := p.Close()
	raws = append(raws, closed...)

	var events []agentbridge.Event
	for _, raw := range raws {
		evs, _, err := Translate(raw)
		if err != nil {
			t.Fatalf("translate: %v", err)
		}
		events = append(events, evs...)
	}
	return events
}

// TestGoldenNormalJSONL feeds normal.jsonl through Parser + Translate.
func TestGoldenNormalJSONL(t *testing.T) {
	events := runFixtureThroughParser(t, "normal.jsonl")
	var saw struct {
		session, lifecycle, text, thinking, toolStart, toolDone, usage, result bool
	}
	for _, ev := range events {
		switch ev.Kind {
		case agentbridge.EventSessionIdentified:
			saw.session = true
		case agentbridge.EventLifecycle:
			saw.lifecycle = true
		case agentbridge.EventTextDelta:
			saw.text = true
		case agentbridge.EventThinkingDelta:
			saw.thinking = true
		case agentbridge.EventToolCallStarted:
			saw.toolStart = true
		case agentbridge.EventToolCallCompleted:
			saw.toolDone = true
		case agentbridge.EventUsageDelta:
			saw.usage = true
		case agentbridge.EventResult:
			if ev.Result.Status == agentbridge.ResultCompleted {
				saw.result = true
			}
		}
	}
	if !saw.session || !saw.lifecycle || !saw.text || !saw.thinking || !saw.toolStart || !saw.toolDone || !saw.usage || !saw.result {
		t.Fatalf("normal.jsonl coverage gap: %+v", saw)
	}
}

// TestGoldenPrefixedLinesJSONL asserts the parser strips stdout:/STDOUT:
// prefixes some wrapper scripts inject and that translation still works.
func TestGoldenPrefixedLinesJSONL(t *testing.T) {
	events := runFixtureThroughParser(t, "prefixed_lines.jsonl")
	if len(events) == 0 {
		t.Fatalf("prefix-stripping fixture produced no events — parser likely refused the prefixed lines")
	}
	gotResult := false
	for _, ev := range events {
		if ev.Kind == agentbridge.EventResult && ev.Result.Status == agentbridge.ResultCompleted {
			gotResult = true
		}
	}
	if !gotResult {
		t.Fatalf("prefixed_lines fixture did not produce a completed Result")
	}
}
