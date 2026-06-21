package supervisor

import (
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorBlocksPreparingTaskOnRuntimeDrift(t *testing.T) {
	cloneStarted, cloneDone := stubCancellableAssignmentClone(t)
	source := &heartbeatDuringPrepareSource{
		req:        slowPrepareTaskRequest(),
		heartbeats: make(chan controlplane.RuntimeHeartbeat, 8),
	}
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	adapter := newMutableDetectAdapter("codex", "1.0.0")
	rt := startRuntimeWithAdapter(t, fake, "rt-codex", adapter)

	startRoutingSupervisor(t, Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		Workdir:        workdir.NewFSAdapter(t.TempDir()),
		HeartbeatEvery: 10 * time.Millisecond,
	})
	expectStartedTask(t, reporter, "asn-slow-prepare")
	expectSignal(t, cloneStarted, "worktree materialization did not start")
	adapter.setVersion("2.0.0")

	res := expectTaskResult(t, reporter, "runtime drift result was not reported")
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("status = %s, want %s", res.Status, agentbridge.ResultBlocked)
	}
	if !strings.Contains(res.Error, runtimeactor.ErrRuntimePinViolated.Error()) {
		t.Fatalf("error = %q", res.Error)
	}
	expectSignal(t, cloneDone, "prepare context was not cancelled after drift")
	assertProcessDoesNotStart(t, running, "provider process started after drift")
}
