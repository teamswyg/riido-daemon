package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func waitPolicyBundleTaskClaimed(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}
}

func completePolicyBundleTask(
	t *testing.T,
	reporter *reporterProbe,
	running *process.FakeRunning,
) agentbridge.Result {
	t.Helper()
	go func() {
		running.EmitStdout([]byte("ok"))
		running.EmitExit(0, nil)
	}()
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
		return res
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}
	return agentbridge.Result{}
}
