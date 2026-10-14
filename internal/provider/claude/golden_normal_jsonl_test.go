package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestGoldenNormalJSONL(t *testing.T) {
	raws := loadGoldenFixtureLines(t, "normal.jsonl")
	if len(raws) == 0 {
		t.Fatal("fixture empty")
	}

	var saw struct {
		session bool
		text    bool
		result  bool
	}
	for _, raw := range raws {
		events, _, err := Translate(raw)
		if err != nil {
			t.Fatalf("translate %+v: %v", raw, err)
		}
		for _, event := range events {
			switch event.Kind {
			case agentbridge.EventSessionIdentified:
				saw.session = true
			case agentbridge.EventTextDelta:
				saw.text = true
			case agentbridge.EventResult:
				saw.result = true
			}
		}
	}
	if !saw.session || !saw.text || !saw.result {
		t.Fatalf("fixture coverage gap: %+v", saw)
	}
}
