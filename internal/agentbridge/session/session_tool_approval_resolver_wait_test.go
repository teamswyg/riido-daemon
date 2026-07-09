package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionResolverWaitCanBeCancelled(t *testing.T) {
	entered := make(chan struct{})
	started := startRecordingSession(t, "assignment-resolver-cancel", approvalAdapter(t, "shell"), func(cfg *Config) {
		cfg.ToolApprovalResolver = waitingResolver(entered)
	})

	started.running.EmitStdout([]byte("ASK"))
	waitResolverEntered(t, entered)
	started.sess.Cancel(errors.New("user stop"))
	res := waitResult(t, started.sess, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("result=%+v", res)
	}
	if res.Error != "user stop" {
		t.Fatalf("cancel reason=%q", res.Error)
	}
}

func TestSessionResolverWaitHonorsHardTimeout(t *testing.T) {
	entered := make(chan struct{})
	started := startRecordingSession(t, "assignment-resolver-timeout", approvalAdapter(t, "shell"), func(cfg *Config) {
		cfg.HardTimeout = 50 * time.Millisecond
		cfg.ToolApprovalResolver = waitingResolver(entered)
	})

	started.running.EmitStdout([]byte("ASK"))
	waitResolverEntered(t, entered)
	res := waitResult(t, started.sess, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("result=%+v", res)
	}
	if res.Error != "hard timeout" {
		t.Fatalf("timeout reason=%q", res.Error)
	}
}

func waitingResolver(entered chan<- struct{}) agentbridge.ToolApprovalResolver {
	return resolverFunc(func(ctx context.Context, _ string, _ agentbridge.ToolRef) (agentbridge.ToolApprovalResolution, error) {
		close(entered)
		<-ctx.Done()
		return agentbridge.ToolApprovalResolution{}, ctx.Err()
	})
}

func waitResolverEntered(t *testing.T, entered <-chan struct{}) {
	t.Helper()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("resolver was not entered")
	}
}
