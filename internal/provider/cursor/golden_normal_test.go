package cursor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestGoldenNormalJSONL(t *testing.T) {
	events := runFixtureThroughParser(t, "normal.jsonl")
	var saw goldenCoverage
	for _, ev := range events {
		recordGoldenEvent(&saw, ev)
	}
	if !saw.session || !saw.lifecycle || !saw.text || !saw.thinking || !saw.toolStart || !saw.toolDone || !saw.usage || !saw.result {
		t.Fatalf("normal.jsonl coverage gap: %+v", saw)
	}
}

type goldenCoverage struct {
	session, lifecycle, text, thinking, toolStart, toolDone, usage, result bool
}

func recordGoldenEvent(saw *goldenCoverage, ev agentbridge.Event) {
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
		saw.result = ev.Result.Status == agentbridge.ResultCompleted
	}
}
