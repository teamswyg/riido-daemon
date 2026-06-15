package session

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionRemovesTempFilesAfterTimeout(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID:      "task-tempfile-timeout",
		RuntimeID:   "rt-1",
		Adapter:     adapter,
		Process:     fake,
		Spawn:       process.Command{Executable: "fake"},
		HardTimeout: 20 * time.Millisecond,
		TempFiles:   []string{tempFile.Name()},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("result: %+v", res)
	}
	if _, err := os.Stat(tempFile.Name()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed after timeout, stat err=%v", err)
	}
	_ = drainEvents(t, sess, time.Second)
}

// Hard timeout fires when no Result arrives in time.
func TestSessionHardTimeout(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			return nil, nil, nil // produce no events ever
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:      "task-2",
		RuntimeID:   "rt-1",
		Adapter:     adapter,
		Process:     fake,
		Spawn:       process.Command{Executable: "fake"},
		HardTimeout: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("status: %s", res.Status)
	}

	// Process must have been killed.
	select {
	case <-sess.runningForTest().KillRecv():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected kill on timeout")
	}
}

// External Cancel beats a later Result event.
func TestSessionCancellation(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(_ agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID: "task-3", RuntimeID: "rt-1", Adapter: adapter,
		Process: fake, Spawn: process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	sess.Cancel(nil)
	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("status: %s", res.Status)
	}
}

// Process exits non-zero without a provider result → failed.
func TestSessionProcessExitNonZero(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(_ agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID: "task-4", RuntimeID: "rt-1", Adapter: adapter,
		Process: fake, Spawn: process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	go sess.runningForTest().EmitExit(137, nil)

	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultFailed {
		t.Fatalf("status: %s", res.Status)
	}
}

// SessionID propagation — first SessionIdentified flows into Result.SessionID.
func TestSessionPropagatesSessionID(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type == "chunk" {
				return []agentbridge.Event{
					{Kind: agentbridge.EventSessionIdentified, SessionID: "sess-abc"},
					{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}},
				}, nil, nil
			}
			return nil, nil, nil
		},
	}
	sess, _ := Start(context.Background(), Config{
		TaskID: "task-5", RuntimeID: "rt-1", Adapter: adapter,
		Process: fake, Spawn: process.Command{Executable: "fake"},
	})
	go sess.runningForTest().EmitStdout([]byte("x"))
	res := waitResult(t, sess, time.Second)
	if res.SessionID != "sess-abc" {
		t.Fatalf("session id: %q", res.SessionID)
	}
}
