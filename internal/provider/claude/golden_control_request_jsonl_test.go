package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestGoldenControlRequestJSONL(t *testing.T) {
	raws := loadGoldenFixtureLines(t, "control_request.jsonl")
	var approval bool
	var failedResult bool
	for _, raw := range raws {
		events, _, _ := Translate(raw)
		for _, event := range events {
			if event.Kind == agentbridge.EventToolApprovalNeeded {
				approval = true
			}
			if event.Kind == agentbridge.EventResult && event.Result.Status == agentbridge.ResultFailed {
				failedResult = true
			}
		}
	}
	if !approval || !failedResult {
		t.Fatalf("control_request coverage gap: approval=%v failed=%v", approval, failedResult)
	}
}
