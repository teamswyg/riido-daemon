package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func expectTaskResultStarted(t *testing.T, reporter *reporterProbe, want string) {
	t.Helper()
	select {
	case taskID := <-reporter.started:
		if taskID != want {
			t.Fatalf("started task: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}
}

func expectTaskResultRunningEvent(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventLifecycle || ev.Phase != agentbridge.StateRunning {
			t.Fatalf("running event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("running event was not reported")
	}
}

func completeTaskResultProcess(
	running interface {
		EmitStdout([]byte)
		EmitExit(int, error)
	},
) {
	go func() {
		running.EmitStdout([]byte("done"))
		running.EmitExit(0, nil)
	}()
}
