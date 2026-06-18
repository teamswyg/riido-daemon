package supervisor

import (
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func expectStopArchiveEvent(
	t *testing.T,
	reporter *reporterProbe,
	timeoutMessage string,
) agentbridge.Event {
	t.Helper()
	select {
	case event := <-reporter.events:
		return event
	case <-time.After(2 * time.Second):
		t.Fatal(timeoutMessage)
	}
	return agentbridge.Event{}
}

func expectStopArchiveProcessKilled(t *testing.T, running *process.FakeRunning) {
	t.Helper()
	select {
	case <-running.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not killed on supervisor stop")
	}
}

func expectStopArchiveCancelledResult(
	t *testing.T,
	reporter *reporterProbe,
	wantWorkdir string,
) agentbridge.Result {
	t.Helper()
	select {
	case result := <-reporter.results:
		if result.Status != agentbridge.ResultCancelled ||
			!strings.Contains(result.Error, "supervisor: stopped") {
			t.Fatalf("shutdown result: %+v", result)
		}
		if result.Workdir != wantWorkdir {
			t.Fatalf("shutdown result workdir = %q, want %q", result.Workdir, wantWorkdir)
		}
		return result
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown result was not reported")
	}
	return agentbridge.Result{}
}
