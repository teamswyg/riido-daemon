package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionNormalCompletion(t *testing.T) {
	started := startRecordingSession(t, "task-1", normalCompletionAdapter(), nil)
	go func() {
		started.running.EmitStdout([]byte("hello"))
		started.running.EmitStdout([]byte("DONE"))
		started.running.EmitExit(0, nil)
	}()

	res := waitResult(t, started.sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("status: %s", res.Status)
	}
	if res.Output != "hello" {
		t.Fatalf("output: %q", res.Output)
	}
	assertTextDeltaSeen(t, drainEvents(t, started.sess, time.Second), "hello")
}

func normalCompletionAdapter() *recordingAdapter {
	return &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type != "chunk" {
				return nil, nil, nil
			}
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{completedResultEvent("hello")}, nil, nil
			}
			return []agentbridge.Event{
				{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning},
				{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)},
			}, nil, nil
		},
	}
}
