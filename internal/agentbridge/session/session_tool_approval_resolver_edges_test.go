package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionResolverErrorFallsBackToApprovalGate(t *testing.T) {
	started := startRecordingSession(t, "assignment-resolver-error", approvalAdapter(t, "shell"), func(cfg *Config) {
		cfg.ToolApprovalResolver = resolverFunc(func(context.Context, string, agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
			return agentbridge.ToolApprovalResolution{}, errors.New("resolver down")
		})
		cfg.ToolApprovalGate = func(agentbridge.ToolRef) agentbridge.ToolStartDecision {
			return agentbridge.ToolStartDecision{Block: true, Reason: "manual approval unavailable"}
		}
	})

	started.running.EmitStdout([]byte("ASK"))
	expectToolProviderKill(t, started.running)
	res := waitResult(t, started.sess, 2*time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result=%+v", res)
	}
	if res.Error != "manual approval unavailable" {
		t.Fatalf("blocked error=%q", res.Error)
	}
	events := drainEvents(t, started.sess, time.Second)
	if !hasWarningText(events, "tool approval resolver failed") {
		t.Fatalf("missing resolver warning in events: %+v", events)
	}
}

func TestSessionResolverDeniedApprovalUsesDefaultReason(t *testing.T) {
	started := startRecordingSession(t, "assignment-denied-default", approvalCommandAdapter(t, "shell", agentbridge.CommandRejectTool), func(cfg *Config) {
		cfg.ToolApprovalResolver = resolverFunc(func(context.Context, string, agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
			return agentbridge.ToolApprovalResolution{}, nil
		})
		cfg.ToolApprovalGate = func(agentbridge.ToolRef) agentbridge.ToolStartDecision {
			t.Fatal("resolver denial should not reach headless gate")
			return agentbridge.ToolStartDecision{}
		}
	})

	started.running.EmitStdout([]byte("ASK"))
	assertRejectionProviderInput(t, started.running)
	started.running.EmitStdout([]byte("DONE"))
	res := waitResult(t, started.sess, 2*time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result=%+v", res)
	}
	if res.Error != "tool approval denied" {
		t.Fatalf("blocked error=%q", res.Error)
	}
	_ = drainEvents(t, started.sess, time.Second)
}
