package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestSupervisorRetriesShutdownTerminalReportUntilAccepted(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(forcedStopRequest())
	reporter := newTerminalRetryReporter(1)
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	actor := startRoutingSupervisor(t, Config{
		DaemonID:           "daemon-shutdown-report-retry",
		RiidoDaemonVersion: "riido-agentd v1.2.3",
		Runtime:            startRuntime(t, fake),
		Source:             source,
		Reporter:           reporter,
		Workdir:            workdir.NewFSAdapter(t.TempDir()),
	})

	expectStartedTask(t, reporter.reporterProbe, "t-forced-stop")
	expectProcessStarted(t, running, "provider process was not spawned")
	shutdownCtx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownForced, 2*time.Second)
	defer cancel()
	if err := actor.StopLifecycle(shutdownCtx); err != nil {
		t.Fatalf("StopLifecycle: %v", err)
	}
	expectTerminalAttempt(t, reporter, 1)
	expectTerminalAttempt(t, reporter, 2)
	expectTaskResultCancelled(t, reporter.reporterProbe)
}

func expectTaskResultCancelled(t *testing.T, reporter *reporterProbe) {
	t.Helper()
	res := expectTaskResult(t, reporter, "shutdown result was not reported")
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("terminal status = %s, want %s", res.Status, agentbridge.ResultCancelled)
	}
}
