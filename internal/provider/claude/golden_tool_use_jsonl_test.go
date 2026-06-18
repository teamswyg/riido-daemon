package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestGoldenToolUseJSONL(t *testing.T) {
	raws := loadGoldenFixtureLines(t, "tool_use.jsonl")
	var started, completed, failed bool
	for _, raw := range raws {
		events, _, err := Translate(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			switch event.Kind {
			case agentbridge.EventToolCallStarted:
				started = true
			case agentbridge.EventToolCallCompleted:
				completed = true
			case agentbridge.EventToolCallFailed:
				failed = true
			}
		}
	}
	if !started || !completed || !failed {
		t.Fatalf("tool_use coverage gap: started=%v completed=%v failed=%v", started, completed, failed)
	}
}
