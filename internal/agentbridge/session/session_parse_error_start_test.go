package session

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type startedEdgeSession struct {
	session *Session
	running *process.FakeRunning
}

func startEdgeSession(t *testing.T, taskID string, adapter agentbridge.Adapter) startedEdgeSession {
	t.Helper()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	sess, err := Start(context.Background(), Config{
		TaskID:    taskID,
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return startedEdgeSession{session: sess, running: running}
}
