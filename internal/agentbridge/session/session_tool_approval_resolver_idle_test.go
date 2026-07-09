package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionResolverWaitHonorsSemanticIdleTimeout(t *testing.T) {
	entered := make(chan struct{})
	started := startRecordingSession(t, "assignment-resolver-idle", approvalAdapter(t, "shell"), func(cfg *Config) {
		cfg.SemanticIdle = 50 * time.Millisecond
		cfg.ToolApprovalResolver = waitingResolver(entered)
	})

	started.running.EmitStdout([]byte("ASK"))
	waitResolverEntered(t, entered)
	res := waitResult(t, started.sess, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("result=%+v", res)
	}
	if res.Error != "semantic idle timeout" {
		t.Fatalf("timeout reason=%q", res.Error)
	}
}
