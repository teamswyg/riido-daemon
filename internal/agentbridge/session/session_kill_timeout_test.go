package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionTimeoutResultDoesNotWaitForeverForBlockingKill(t *testing.T) {
	proc := &blockingKillProcess{running: newBlockingKillRunning()}
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID:             "task-blocking-kill-timeout",
		RuntimeID:          "rt-1",
		Adapter:            adapter,
		Process:            proc,
		Spawn:              process.Command{Executable: "fake"},
		HardTimeout:        10 * time.Millisecond,
		ProcessKillTimeout: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("result: %+v", res)
	}
	select {
	case <-proc.running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("session did not request provider kill")
	}
}
