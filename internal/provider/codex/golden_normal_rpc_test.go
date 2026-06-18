package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestGoldenNormalRPC(t *testing.T) {
	raws := loadGoldenFixtureLines(t, "normal_rpc.jsonl")
	var saw normalRPCCoverage
	for _, raw := range raws {
		events, _, err := Translate(raw)
		if err != nil {
			t.Fatal(err)
		}
		observeNormalRPCEvent(&saw, events)
	}
	if !saw.session || !saw.lifecycle || !saw.text || !saw.reasoning ||
		!saw.toolStart || !saw.toolDone || !saw.usage || !saw.result {
		t.Fatalf("normal_rpc coverage gap: %+v", saw)
	}
}

type normalRPCCoverage struct {
	session   bool
	lifecycle bool
	text      bool
	reasoning bool
	toolStart bool
	toolDone  bool
	usage     bool
	result    bool
}

func observeNormalRPCEvent(saw *normalRPCCoverage, events []agentbridge.Event) {
	for _, event := range events {
		switch event.Kind {
		case agentbridge.EventSessionIdentified:
			saw.session = true
		case agentbridge.EventLifecycle:
			saw.lifecycle = true
		case agentbridge.EventTextDelta:
			saw.text = true
		case agentbridge.EventThinkingDelta:
			saw.reasoning = true
		case agentbridge.EventToolCallStarted:
			saw.toolStart = true
		case agentbridge.EventToolCallCompleted:
			saw.toolDone = true
		case agentbridge.EventUsageDelta:
			saw.usage = true
		case agentbridge.EventResult:
			if event.Result.Status == agentbridge.ResultCompleted {
				saw.result = true
			}
		}
	}
}
