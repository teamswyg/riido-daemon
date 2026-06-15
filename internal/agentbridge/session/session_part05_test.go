package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Result is emitted exactly once on the Result channel.
func TestSessionResultExactlyOnce(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type == "chunk" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			}
			return nil, nil, nil
		},
	}
	sess, _ := Start(context.Background(), Config{
		TaskID: "task-6", RuntimeID: "rt-1", Adapter: adapter,
		Process: fake, Spawn: process.Command{Executable: "fake"},
	})

	go func() {
		sess.runningForTest().EmitStdout([]byte("x"))
		sess.runningForTest().EmitStdout([]byte("y"))
		sess.runningForTest().EmitExit(0, nil)
	}()

	_ = waitResult(t, sess, time.Second)
	// Second read MUST observe a closed channel (no second result).
	select {
	case _, ok := <-sess.Result():
		if ok {
			t.Fatal("Result channel must close after one delivery")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Result channel didn't close")
	}
}
