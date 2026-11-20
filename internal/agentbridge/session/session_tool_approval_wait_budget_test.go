package session

import (
	"context"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionApprovalWaitPreservesRemainingExecutionBudget(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		requested, decision := make(chan struct{}), make(chan struct{})
		started := startRecordingSession(t, "assignment-1", approvalAdapter(t, "patch_apply"), func(cfg *Config) {
			cfg.SemanticIdle = time.Second
			cfg.HardTimeout = 100 * time.Millisecond
			cfg.ToolApprovalResolver = resolverFunc(func(ctx context.Context, _ string, _ agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
				close(requested)
				select {
				case <-decision:
					return agentbridge.ToolApprovalResolution{Approved: true}, nil
				case <-ctx.Done():
					return agentbridge.ToolApprovalResolution{}, ctx.Err()
				}
			})
		})
		t.Cleanup(func() { started.sess.Cancel(nil) })
		time.Sleep(40 * time.Millisecond)
		started.running.EmitStdout([]byte("ASK"))
		<-requested
		time.Sleep(24 * time.Hour)
		synctest.Wait()
		select {
		case res := <-started.sess.Result():
			t.Fatalf("human response time consumed execution budget: %+v", res)
		default:
		}
		select {
		case frame := <-started.running.StdinRecv():
			t.Fatalf("provider input before explicit approval: %q", frame)
		default:
		}
		close(decision)
		assertApprovalProviderInput(t, started.running)
		synctest.Wait()
		time.Sleep(59 * time.Millisecond)
		select {
		case res := <-started.sess.Result():
			t.Fatalf("remaining execution budget ended early: %+v", res)
		default:
		}
		time.Sleep(time.Millisecond)
		synctest.Wait()
		res := waitResult(t, started.sess, time.Second)
		if res.Status != agentbridge.ResultTimeout || !strings.Contains(res.Error, "hard timeout") {
			t.Fatalf("active execution budget was not enforced: %+v", res)
		}
		events := drainEvents(t, started.sess, time.Second)
		resumed := false
		for _, ev := range events {
			if ev.Kind == agentbridge.EventLifecycle && ev.Phase == agentbridge.StateRunning {
				resumed = true
			}
		}
		if !resumed {
			t.Fatal("explicit approval did not restore running status")
		}
	})
}
