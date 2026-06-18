package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionAutoApprovalWritesProviderInput(t *testing.T) {
	started := startRecordingSession(t, "task-approval", approvalAdapter(t, "read"), func(cfg *Config) {
		cfg.AutoApprove = func(tool agentbridge.ToolRef) bool { return tool.Kind == "read" }
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
