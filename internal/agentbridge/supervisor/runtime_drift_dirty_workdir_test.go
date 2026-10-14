package supervisor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSupervisorDoesNotReinjectDirtyRunningWorkdirOnRuntimeDrift(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(dirtyWorkdirDriftRequest())
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	adapter := newMutableDetectAdapter("fake", "1.0.0")
	rt := startRuntimeWithAdapter(t, fake, "rt-local", adapter)
	workdirs := newCountingWorkdir(t.TempDir())

	startRoutingSupervisor(t, Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		Workdir:        workdirs,
		HeartbeatEvery: 10 * time.Millisecond,
	})
	expectStartedTask(t, reporter, "t-dirty-drift")
	expectProviderStarted(t, running)
	makeWorkdirDirty(t, workdirs)
	adapter.setVersion("2.0.0")

	res := expectTaskResult(t, reporter, "runtime drift result was not reported")
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("status = %s, want %s", res.Status, agentbridge.ResultBlocked)
	}
	if !strings.Contains(res.Error, runtimeactor.ErrRuntimePinViolated.Error()) {
		t.Fatalf("error = %q", res.Error)
	}
	_, injects := workdirs.snapshot()
	if injects != 1 {
		t.Fatalf("InjectRuntimeConfig calls = %d, want 1", injects)
	}
}

func dirtyWorkdirDriftRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:       "t-dirty-drift",
		Provider: "fake",
		Prompt:   "run",
		Metadata: map[string]string{MetadataWorkspaceID: "ws-dirty"},
	}
}

func makeWorkdirDirty(t *testing.T, workdirs *countingWorkdir) {
	t.Helper()
	ws, _ := workdirs.snapshot()
	if err := os.WriteFile(filepath.Join(ws.Workdir, "dirty.txt"), []byte("dirty"), 0o600); err != nil {
		t.Fatalf("write dirty marker: %v", err)
	}
}
