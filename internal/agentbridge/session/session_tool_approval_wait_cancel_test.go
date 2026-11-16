package session

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionResolverApprovalWaitRemainsCancellable(t *testing.T) {
	for _, cancelContext := range []bool{false, true} {
		t.Run(map[bool]string{false: "session", true: "context"}[cancelContext], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				requested, resolverDone := make(chan struct{}), make(chan struct{})
				started := startRecordingSessionContext(t, ctx, "assignment-1", approvalAdapter(t, "patch_apply"), func(cfg *Config) {
					cfg.SemanticIdle = 100 * time.Millisecond
					cfg.ToolApprovalResolver = resolverFunc(func(ctx context.Context, _ string, _ agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
						defer close(resolverDone)
						close(requested)
						<-ctx.Done()
						return agentbridge.ToolApprovalResolution{}, ctx.Err()
					})
				})
				t.Cleanup(func() { started.sess.Cancel(nil) })
				started.running.EmitStdout([]byte("ASK"))
				<-requested
				time.Sleep(200 * time.Millisecond)
				if cancelContext {
					cancel()
				} else {
					started.sess.Cancel(nil)
				}
				res := waitResult(t, started.sess, time.Second)
				if res.Status != agentbridge.ResultCancelled {
					t.Fatalf("result: %+v", res)
				}
				<-resolverDone
				select {
				case frame := <-started.running.StdinRecv():
					t.Fatalf("sent a decision after cancellation: %q", frame)
				default:
				}
				_ = drainEvents(t, started.sess, time.Second)
			})
		})
	}
}
