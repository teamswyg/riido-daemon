package supervisor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestSupervisorRetriesRuntimeDeregisterUntilAccepted(t *testing.T) {
	source := newRuntimeDeregisterRetrySource(1)
	actor := startRoutingSupervisor(t, Config{
		DaemonID: "daemon-deregister-retry",
		Runtime:  startRuntime(t, process.NewFake()),
		Source:   source,
		Reporter: controlplane.NewMemoryReporter(),
		Workdir:  workdir.NewFSAdapter(t.TempDir()),
	})

	shutdownCtx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownGraceful, 2*time.Second)
	defer cancel()
	if err := actor.StopLifecycle(shutdownCtx); err != nil {
		t.Fatalf("StopLifecycle: %v", err)
	}
	expectDeregisterAttempt(t, source, 1)
	expectDeregisterAttempt(t, source, 2)
	expectDeregisteredRuntime(t, source.runtimeRoutingSource, "rt-local")
}

func expectDeregisteredRuntime(t *testing.T, source *runtimeRoutingSource, want string) {
	t.Helper()
	select {
	case got := <-source.deregistered:
		if got != want {
			t.Fatalf("deregistered runtime = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runtime deregistration was not accepted")
	}
}
