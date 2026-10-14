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

func TestSessionResolverDeniedApprovalBlocksProviderCompletion(t *testing.T) {
	started := startRecordingSession(t, "assignment-1", approvalCommandAdapter(t, "shell", agentbridge.CommandRejectTool), func(cfg *Config) {
		cfg.ToolApprovalGate = func(agentbridge.ToolRef) agentbridge.ToolStartDecision {
			t.Fatal("resolver-denied request should not reach headless gate")
			return agentbridge.ToolStartDecision{}
		}
		cfg.ToolApprovalResolver = resolverFunc(func(_ context.Context, executionID string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
			if executionID != "assignment-1" || tool.ID != "tool-1" {
				t.Fatalf("resolver input execution=%s tool=%+v", executionID, tool)
			}
			return agentbridge.ToolApprovalResolution{Reason: "tool approval timed out"}, nil
		})
	})

	started.running.EmitStdout([]byte("ASK"))
	assertRejectionProviderInput(t, started.running)
	started.running.EmitStdout([]byte("DONE"))
	res := waitResult(t, started.sess, 2*time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result: %+v", res)
	}
	if res.Error != "tool approval timed out" {
		t.Fatalf("blocked error: %q", res.Error)
	}
	_ = drainEvents(t, started.sess, time.Second)
}
