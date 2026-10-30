package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestIdleUpdateIncludesPreparationAndUndeliveredReports(t *testing.T) {
	for _, tc := range []struct {
		name     string
		flight   map[string]*runningTask
		detached []detachedReport
		mailbox  int
	}{
		{name: "preparing", flight: map[string]*runningTask{"preparing": {}}},
		{name: "terminal awaiting ACK", flight: map[string]*runningTask{"terminal": {}}},
		{name: "detached report", detached: []detachedReport{{}}},
		{name: "queued mailbox", mailbox: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if idleForUpdate(tc.flight, tc.detached, tc.mailbox) {
				t.Fatal("unsafe update accepted")
			}
		})
	}
	if !idleForUpdate(nil, nil, 0) {
		t.Fatal("idle update rejected")
	}
}

func TestIdleUpdateFencesNewClaimsAtomically(t *testing.T) {
	source := newIdlePollSource()
	actor := startRoutingSupervisor(t, Config{DaemonID: "daemon-1", Runtime: startRuntime(t, process.NewFake()), Source: source, Reporter: newReporterProbe(), Workdir: workdir.NewFSAdapter(t.TempDir()), PollEvery: 5 * time.Millisecond, IdlePollEvery: 5 * time.Millisecond})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if !actor.PrepareIdleUpdate(ctx) {
		t.Fatal("idle supervisor refused fence")
	}
	for len(source.claims) > 0 {
		<-source.claims
	}
	select {
	case <-source.claims:
		t.Fatal("new claim raced accepted update")
	case <-time.After(50 * time.Millisecond):
	}
}

type silentNotifyingSource struct {
	*controlplane.MemorySource
	notices chan struct{}
}

func (s *silentNotifyingSource) AssignmentNotifications(context.Context) <-chan struct{} {
	return s.notices
}

func TestIdleUpdateDefersThroughExecutionAndTerminalRetryWithoutNotifications(t *testing.T) {
	source := &silentNotifyingSource{MemorySource: controlplane.NewMemorySource(), notices: make(chan struct{})}
	source.Enqueue(taskResultRequest())
	reporter := newTerminalRetryReporter(1)
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	actor := startRoutingSupervisor(t, Config{DaemonID: "daemon-1", Runtime: startRuntime(t, fake), Source: source, Reporter: reporter, Workdir: workdir.NewFSAdapter(t.TempDir()), PollEvery: 20 * time.Millisecond})
	expectTaskResultStarted(t, reporter.reporterProbe, "t-1")
	expectTaskResultRunningEvent(t, reporter.reporterProbe)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if actor.PrepareIdleUpdate(ctx) {
		t.Fatal("running execution was interrupted")
	}
	completeTaskResultProcess(running)
	expectTerminalAttempt(t, reporter, 1)
	if actor.PrepareIdleUpdate(ctx) {
		t.Fatal("unacknowledged terminal report was discarded")
	}
	expectTerminalAttempt(t, reporter, 2)
	expectTaskResultCompleted(t, reporter.reporterProbe)
	if !actor.PrepareIdleUpdate(ctx) {
		t.Fatal("completed and acknowledged run prevented update")
	}
}
