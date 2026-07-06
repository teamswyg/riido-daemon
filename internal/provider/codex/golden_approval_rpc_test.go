package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestGoldenApprovalRPC(t *testing.T) {
	raws := loadGoldenFixtureLines(t, "approval_rpc.jsonl")
	approvalCommands := 0
	failedResult := false
	for _, raw := range raws {
		events, cmds, _ := Translate(raw)
		for _, cmd := range cmds {
			if cmd.Kind == agentbridge.CommandApproveTool {
				approvalCommands++
			}
		}
		for _, event := range events {
			if event.Kind == agentbridge.EventResult && event.Result.Status == agentbridge.ResultFailed {
				failedResult = true
			}
		}
	}
	if approvalCommands != 2 {
		t.Fatalf("expected 2 approval commands (command + patch), got %d", approvalCommands)
	}
	if !failedResult {
		t.Fatalf("approval_rpc fixture missing failed result event")
	}
}
