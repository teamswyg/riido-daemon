package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func expectTaskResultTextDelta(t *testing.T, reporter *reporterProbe, want string) {
	t.Helper()
	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventTextDelta || ev.Text != want {
			t.Fatalf("event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("event was not reported")
	}
}

func expectTaskResultCompletedRun(t *testing.T, run taskResultSupervisorRun) {
	t.Helper()
	select {
	case res := <-run.reporter.results:
		assertSupervisorCompletedRun(t, res, run.running)
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}
}
