package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRequiredHumanApprovalWaitsEvenWithAutoApprovalEnabled(t *testing.T) {
	const kind = "mcp_confirmation"
	adapter := approvalAdapter(t, kind)
	translate := adapter.translateFn
	adapter.translateFn = func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
		events, commands, err := translate(raw)
		if len(events) == 1 && events[0].Kind == agentbridge.EventToolApprovalNeeded {
			events[0].Tool.RequireHuman = true
		}
		return events, commands, err
	}
	requested := make(chan agentbridge.ToolRef, 1)
	approved := make(chan struct{})
	started := startRecordingSession(t, "assignment-1", adapter, func(cfg *Config) {
		cfg.AutoApprove = func(agentbridge.ToolRef) bool { return true }
		cfg.ToolApprovalResolver = resolverFunc(func(ctx context.Context, _ string, tool agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
			requested <- tool
			select {
			case <-approved:
				return agentbridge.ToolApprovalResolution{Approved: true}, nil
			case <-ctx.Done():
				return agentbridge.ToolApprovalResolution{}, ctx.Err()
			}
		})
	})
	started.running.EmitStdout([]byte("ASK"))
	select {
	case tool := <-requested:
		if tool.Kind != kind || !tool.RequireHuman {
			t.Fatalf("unexpected confirmation: %+v", tool)
		}
	case <-time.After(time.Second):
		t.Fatal("approval resolver was not reached")
	}
	select {
	case frame := <-started.running.StdinRecv():
		t.Fatalf("decision sent before human approval: %q", frame)
	default:
	}
	close(approved)
	assertApprovalProviderInput(t, started.running)
	emitDone(started.running)
	if res := waitResult(t, started.sess, 2*time.Second); res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	_ = drainEvents(t, started.sess, time.Second)
}

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

func TestSessionResolverSemanticIdleDoesNotGrantApproval(t *testing.T) {
	requested := make(chan struct{})
	started := startRecordingSession(t, "assignment-1", approvalAdapter(t, "patch_apply"), func(cfg *Config) {
		cfg.SemanticIdle = 150 * time.Millisecond
		cfg.ToolApprovalResolver = resolverFunc(func(ctx context.Context, _ string, _ agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
			close(requested)
			<-ctx.Done()
			return agentbridge.ToolApprovalResolution{}, ctx.Err()
		})
	})
	started.running.EmitStdout([]byte("ASK"))
	select {
	case <-requested:
	case <-time.After(time.Second):
		t.Fatal("approval resolver was not reached")
	}
	res := waitResult(t, started.sess, time.Second)
	if res.Status != agentbridge.ResultTimeout || res.Error != "semantic idle timeout" {
		t.Fatalf("result: %+v", res)
	}
	select {
	case frame := <-started.running.StdinRecv():
		t.Fatalf("sent a provider approval decision after idle: %q", frame)
	default:
	}
}
