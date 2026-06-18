package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type cancellationFixture struct {
	source   *cancelSource
	reporter *reporterProbe
	running  *process.FakeRunning
	runtime  *runtimeactor.Actor
}

func newCancellationFixture(t *testing.T, req bridge.TaskRequest, runtimeID, provider string) cancellationFixture {
	t.Helper()
	source := &cancelSource{req: req, cancel: make(chan error, 1)}
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startNamedRuntime(t, fake, runtimeID, provider)
	return cancellationFixture{source: source, reporter: reporter, running: running, runtime: rt}
}

func expectProviderStarted(t *testing.T, running *process.FakeRunning) {
	t.Helper()
	select {
	case <-running.StartedRecv():
	case <-time.After(supervisorCancellationTestTimeout):
		t.Fatal("provider process was not started")
	}
}
