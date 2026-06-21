package session

import (
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionTerminateWithBlockedResult(t *testing.T) {
	started := startRecordingSession(t, "task-terminal", noEventRecordingAdapter(), nil)

	started.sess.TerminateWithContext(t.Context(), agentbridge.Result{
		Status: agentbridge.ResultBlocked,
		Error:  "runtime pin violated",
	})

	res := waitResult(t, started.sess, time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("status = %s, want %s", res.Status, agentbridge.ResultBlocked)
	}
	if !strings.Contains(res.Error, "runtime pin violated") {
		t.Fatalf("error = %q", res.Error)
	}
	select {
	case <-started.running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("terminal result did not kill provider process")
	}
}
