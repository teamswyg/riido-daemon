package session

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type toolGateScenario struct {
	session *Session
	running *process.FakeRunning
}

func startToolGateScenario(
	t *testing.T,
	taskID string,
	adapter *recordingAdapter,
	configure func(*Config),
) toolGateScenario {
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
	return toolGateScenario{session: sess, running: running}
}

func expectToolProviderKill(t *testing.T, running *process.FakeRunning) {
	t.Helper()
	select {
	case <-running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("expected provider kill after tool gate block")
	}
}

func hasWarningText(events []agentbridge.Event, text string) bool {
	for _, ev := range events {
		if ev.Kind == agentbridge.EventWarning && ev.Text == text {
			return true
		}
	}
	return false
}

func newSessionTempFilePath(t *testing.T) string {
	t.Helper()
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return tempFile.Name()
}
