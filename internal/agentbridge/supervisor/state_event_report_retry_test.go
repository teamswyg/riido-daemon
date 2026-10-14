package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorRetriesStateEventBeforeTerminalReport(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(taskResultRequest())
	reporter := newStateEventRetryReporter(1)
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)
	actor, err := New(Config{
		DaemonID:           "daemon-1",
		RiidoDaemonVersion: testRiidoDaemonVersion,
		Runtime:            rt,
		Source:             source,
		Reporter:           reporter,
		Workdir:            workdir.NewFSAdapter(t.TempDir()),
		PollEvery:          10 * time.Millisecond,
		HeartbeatEvery:     time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("supervisor Start: %v", err)
	}
	t.Cleanup(func() { _ = actor.Stop(context.Background()) })

	expectTaskResultStarted(t, reporter.reporterProbe, "t-1")
	expectStateEventAttempt(t, reporter, 1)
	completeTaskResultProcess(running)
	expectStateEventAttempt(t, reporter, 2)
	expectRetriedRunningEvent(t, reporter.reporterProbe)
	expectTaskResultCompleted(t, reporter.reporterProbe)
}
