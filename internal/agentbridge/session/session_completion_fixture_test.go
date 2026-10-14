package session

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type startedRecordingSession struct {
	sess    *Session
	running *process.FakeRunning
}

func startRecordingSession(t *testing.T, taskID string, adapter *recordingAdapter, configure func(*Config)) startedRecordingSession {
	t.Helper()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	cfg := Config{
		TaskID:    taskID,
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
	}
	if configure != nil {
		configure(&cfg)
	}
	sess, err := Start(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return startedRecordingSession{sess: sess, running: running}
}

func emitDone(running *process.FakeRunning) {
	go func() {
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()
}

func completedResultEvent(output string) agentbridge.Event {
	return agentbridge.Event{
		Kind:   agentbridge.EventResult,
		Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: output},
	}
}
