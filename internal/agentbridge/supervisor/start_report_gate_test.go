package supervisor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSupervisorDoesNotStartProviderWhenStartReportFails(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(dirtyWorkdirDriftRequest())
	reporter := newStartFailReporter()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)

	startRoutingSupervisor(t, Config{
		DaemonID: "daemon-1",
		Runtime:  rt,
		Source:   source,
		Reporter: reporter,
	})

	expectStartReportAttempt(t, reporter, "t-dirty-drift")
	assertProcessDoesNotStart(t, running, "provider process started after StartTask failed")
}

type startFailReporter struct {
	attempted chan string
}

func newStartFailReporter() *startFailReporter {
	return &startFailReporter{attempted: make(chan string, 1)}
}

func (r *startFailReporter) StartTask(_ context.Context, taskID string) error {
	r.attempted <- taskID
	return errors.New("start report rejected")
}

func (r *startFailReporter) ReportEvent(context.Context, string, agentbridge.Event) error {
	return nil
}

func (r *startFailReporter) CompleteTask(context.Context, string, agentbridge.Result) error {
	return nil
}
