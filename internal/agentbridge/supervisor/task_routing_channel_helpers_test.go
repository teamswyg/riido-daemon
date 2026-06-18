package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func expectStartedTask(t *testing.T, reporter *reporterProbe, want string) {
	t.Helper()
	select {
	case taskID := <-reporter.started:
		if taskID != want {
			t.Fatalf("started execution: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}
}

func expectTaskResult(t *testing.T, reporter *reporterProbe, timeout string) agentbridge.Result {
	t.Helper()
	select {
	case res := <-reporter.results:
		return res
	case <-time.After(2 * time.Second):
		t.Fatal(timeout)
		return agentbridge.Result{}
	}
}

func completeFakeProcess(running *process.FakeRunning) {
	go func() {
		running.EmitStdout([]byte("done"))
		running.EmitExit(0, nil)
	}()
}

func expectProcessStarted(t *testing.T, running *process.FakeRunning, msg string) {
	t.Helper()
	select {
	case <-running.StartedRecv():
	case <-time.After(2 * time.Second):
		t.Fatal(msg)
	}
}

func expectProcessKilled(t *testing.T, running *process.FakeRunning, msg string) {
	t.Helper()
	select {
	case <-running.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal(msg)
	}
}

func assertProcessDoesNotStart(t *testing.T, running *process.FakeRunning, msg string) {
	t.Helper()
	select {
	case <-running.StartedRecv():
		t.Fatal(msg)
	case <-time.After(100 * time.Millisecond):
	}
}
