package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestGoldenApprovalRPC(t *testing.T) {
	raws := loadGoldenFixtureLines(t, "approval_rpc.jsonl")
	approvals := 0
	failedResult := false
	for _, raw := range raws {
		events, _, _ := Translate(raw)
		for _, event := range events {
			if event.Kind == agentbridge.EventToolApprovalNeeded {
				approvals++
			}
			if event.Kind == agentbridge.EventResult && event.Result.Status == agentbridge.ResultFailed {
				failedResult = true
			}
		}
	}
	if approvals != 2 {
		t.Fatalf("expected 2 approval events (command + patch), got %d", approvals)
	}
	if !failedResult {
		t.Fatalf("approval_rpc fixture missing failed result event")
	}
}
