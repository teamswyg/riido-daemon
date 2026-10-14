package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func waitForNativeConfigTaskClaim(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}
}

func waitForNativeConfigProcessSpawn(t *testing.T, running *process.FakeRunning) {
	t.Helper()
	select {
	case <-running.StartedRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not spawned")
	}
}

func completeNativeConfigProcess(running *process.FakeRunning) {
	running.EmitStdout([]byte("ok"))
	running.EmitExit(0, nil)
}

func waitForNativeConfigResult(t *testing.T, reporter *reporterProbe) agentbridge.Result {
	t.Helper()
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
		return res
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
		return agentbridge.Result{}
	}
}
