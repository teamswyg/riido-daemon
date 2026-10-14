package cursor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestGoldenPrefixedLinesJSONL(t *testing.T) {
	events := runFixtureThroughParser(t, "prefixed_lines.jsonl")
	if len(events) == 0 {
		t.Fatalf("prefix-stripping fixture produced no events")
	}
	if !hasCompletedResult(events) {
		t.Fatalf("prefixed_lines fixture did not produce a completed Result")
	}
}

func hasCompletedResult(events []agentbridge.Event) bool {
	for _, ev := range events {
		if ev.Kind == agentbridge.EventResult && ev.Result.Status == agentbridge.ResultCompleted {
			return true
		}
	}
	return false
}
