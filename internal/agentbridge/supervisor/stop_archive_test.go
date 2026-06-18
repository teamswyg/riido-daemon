package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSupervisorStopArchivesInFlightWorkspace(t *testing.T) {
	fixture := startStopArchiveFixture(t)
	expectStartedTask(t, fixture.reporter, stopArchiveTaskID)
	cmd := expectStopArchiveProviderStarted(t, fixture.running)

	fixture.running.EmitStdout([]byte("event"))
	expectStopArchiveRunningEvent(t, fixture.reporter)
	expectStopArchiveTextEvent(t, fixture.reporter)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := fixture.actor.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	expectStopArchiveProcessKilled(t, fixture.running)
	res := expectStopArchiveCancelledResult(t, fixture.reporter, cmd.Dir)
	assertStopArchiveManifest(t, res.Workdir)
	assertStopArchiveRunEvents(t, res.Workdir)
}

func expectStopArchiveProviderStarted(t *testing.T, running *process.FakeRunning) process.Command {
	t.Helper()
	select {
	case cmd := <-running.StartedRecv():
		if cmd.Dir == "" {
			t.Fatalf("provider command missing workdir: %+v", cmd)
		}
		return cmd
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not spawned")
	}
	return process.Command{}
}

func expectStopArchiveRunningEvent(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	event := expectStopArchiveEvent(t, reporter, "running event was not reported")
	if event.Kind != agentbridge.EventLifecycle || event.Phase != agentbridge.StateRunning {
		t.Fatalf("running event: %+v", event)
	}
}

func expectStopArchiveTextEvent(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	event := expectStopArchiveEvent(t, reporter, "nonterminal event was not reported")
	if event.Kind != agentbridge.EventTextDelta || event.Text != "event" {
		t.Fatalf("event: %+v", event)
	}
}
