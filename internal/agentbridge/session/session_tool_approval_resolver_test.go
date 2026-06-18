package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionResolverApprovalWritesProviderInput(t *testing.T) {
	started := startRecordingSession(t, "assignment-1", approvalAdapter(t, "patch_apply"), func(cfg *Config) {
		cfg.ToolApprovalGate = func(agentbridge.ToolRef) agentbridge.ToolStartDecision {
			t.Fatal("resolver-approved request should not reach headless gate")
			return agentbridge.ToolStartDecision{}
		}
		cfg.ToolApprovalResolver = resolverFunc(func(_ context.Context, executionID string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
			if executionID != "assignment-1" || tool.ID != "tool-1" {
				t.Fatalf("resolver input execution=%s tool=%+v", executionID, tool)
			}
			return agentbridge.ToolApprovalResolution{Approved: true}, nil
		})
	})

	started.running.EmitStdout([]byte("ASK"))
	assertApprovalProviderInput(t, started.running)
	emitDone(started.running)
	res := waitResult(t, started.sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	_ = drainEvents(t, started.sess, time.Second)
}
