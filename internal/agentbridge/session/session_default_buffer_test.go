package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionDefaultBuffersMatchProviderRuntimeBackpressureSSOT(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			}
			return nil, nil, nil
		},
	}

	sess, err := Start(context.Background(), Config{
		TaskID:    "task-default-buffer",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := cap(sess.Events()); got != DefaultEventBuffer {
		t.Fatalf("default event buffer = %d, want %d", got, DefaultEventBuffer)
	}
	if got := cap(sess.Result()); got != DefaultResultBuffer {
		t.Fatalf("default result buffer = %d, want %d", got, DefaultResultBuffer)
	}

	go func() {
		sess.runningForTest().EmitStdout([]byte("DONE"))
		sess.runningForTest().EmitExit(0, nil)
	}()
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	_ = drainEvents(t, sess, time.Second)
}
