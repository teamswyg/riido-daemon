package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestSupervisorStopLifecyclePropagatesForcedLevel(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(forcedStopRequest())
	reporter := newLifecycleReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	actor := startRoutingSupervisor(t, Config{
		DaemonID:           "daemon-forced-stop",
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
	expectLifecycleCompleteLevel(t, reporter, lifecycle.ShutdownForced)
	expectProcessKilled(t, running, "provider process was not killed on forced supervisor stop")
}

func forcedStopRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:       "t-forced-stop",
		Provider: "fake",
		Prompt:   "x",
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-forced-stop",
			MetadataRunID:       "run-forced-stop",
		},
	}
}

func expectLifecycleCompleteLevel(
	t *testing.T,
	reporter *lifecycleReporterProbe,
	want lifecycle.ShutdownLevel,
) {
	t.Helper()
	select {
	case level := <-reporter.completeLevels:
		if level != want {
			t.Fatalf("CompleteTask lifecycle level = %s, want %s", level, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown result was not reported")
	}
}
