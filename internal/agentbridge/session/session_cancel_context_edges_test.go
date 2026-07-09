package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestSessionCancelWithNilContextPreservesCause(t *testing.T) {
	started := startRecordingSession(t, "task-cancel-nil-context", noEventRecordingAdapter(), nil)
	started.sess.CancelWithContext(nil, errors.New("user stopped"))
	res := waitResult(t, started.sess, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("status=%s", res.Status)
	}
	if res.Error != "user stopped" {
		t.Fatalf("error=%q", res.Error)
	}
}

func TestSessionForcedCancelRequestsProviderKill(t *testing.T) {
	proc := &blockingKillProcess{running: newBlockingKillRunning()}
	sess, err := Start(context.Background(), Config{
		TaskID:             "task-forced-cancel",
		RuntimeID:          "rt-1",
		Adapter:            noEventRecordingAdapter(),
		Process:            proc,
		Spawn:              process.Command{Executable: "fake"},
		ProcessKillTimeout: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	base, release := context.WithCancel(context.Background())
	defer release()
	ctx := lifecycle.New(base, lifecycle.ShutdownForced).Context()
	sess.CancelWithContext(ctx, errors.New("forced stop"))
	select {
	case <-proc.running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("forced cancel did not request provider kill")
	}
	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("status=%s", res.Status)
	}
	if res.Error != "forced stop" {
		t.Fatalf("error=%q", res.Error)
	}
	release()
}
