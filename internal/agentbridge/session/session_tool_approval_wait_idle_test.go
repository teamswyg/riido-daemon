package session

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionResolverApprovalWaitPausesIdleAndRestartsAfterDecision(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const idle = 100 * time.Millisecond
		requested := make(chan struct{})
		decision := make(chan struct{})
		started := startRecordingSession(t, "assignment-1", approvalAdapter(t, "patch_apply"), func(cfg *Config) {
			cfg.SemanticIdle = idle
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
		started.running.EmitStdout([]byte("ASK"))
		<-requested
		time.Sleep(2 * idle)
		synctest.Wait()
		select {
		case res := <-started.sess.Result():
			t.Fatalf("approval wait ended on semantic idle: %+v", res)
		default:
		}
		select {
		case frame := <-started.running.StdinRecv():
			t.Fatalf("approval wait granted provider input: %q", frame)
		default:
		}
		close(decision)
		if frame := <-started.running.StdinRecv(); string(frame) != "approve:req-1\n" {
			t.Fatalf("approval frame: %q", frame)
		}
		synctest.Wait()
		time.Sleep(idle - time.Millisecond)
		select {
		case res := <-started.sess.Result():
			t.Fatalf("original idle budget ended early: %+v", res)
		default:
		}
		time.Sleep(time.Millisecond)
		synctest.Wait()
		res := waitResult(t, started.sess, time.Second)
		if res.Status != agentbridge.ResultTimeout || res.Error != "semantic idle timeout" {
			t.Fatalf("provider idle after explicit approval: %+v", res)
		}
		_ = drainEvents(t, started.sess, time.Second)
	})
}
