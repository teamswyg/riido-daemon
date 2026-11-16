package session

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionApprovalWaitStopsWhenProviderDisconnects(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		requested := make(chan struct{})
		started := startRecordingSession(t, "assignment-1", approvalAdapter(t, "patch_apply"), func(cfg *Config) {
			cfg.HardTimeout = time.Second
			cfg.ToolApprovalResolver = resolverFunc(func(ctx context.Context, _ string, _ agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
				close(requested)
				<-ctx.Done()
				return agentbridge.ToolApprovalResolution{}, ctx.Err()
			})
		})
		t.Cleanup(func() { started.sess.Cancel(nil) })
		started.running.EmitStdout([]byte("ASK"))
		<-requested
		go started.running.EmitExit(1, nil)
		synctest.Wait()
		select {
		case res := <-started.sess.Result():
			if res.Status != agentbridge.ResultAborted {
				t.Fatalf("disconnected approval result: %+v", res)
			}
		default:
			t.Fatal("approval wait retained a disconnected provider")
		}
	})
}
