package supervisor

import (
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorCancellationDuringWorkspacePrepareStopsBeforeRuntimeStart(t *testing.T) {
	cloneStarted, cloneCanceled := stubCancellableAssignmentClone(t)
	fixture := newCancellationFixture(t, prepareCancellationRequest(), "rt-codex", "codex")
	startRoutingSupervisor(t, Config{
		DaemonID:       "daemon-1",
		Runtime:        fixture.runtime,
		Source:         fixture.source,
		Reporter:       fixture.reporter,
		Workdir:        workdir.NewFSAdapter(t.TempDir()),
		HeartbeatEvery: time.Second,
	})
	expectStartedTask(t, fixture.reporter, "t-cancel-prepare")
	expectSignal(t, cloneStarted, "worktree materialization did not start")
	fixture.source.cancel <- errors.New("human cancel during prepare")
	expectSignal(t, cloneCanceled, "workspace materialization context was not cancelled")
	res := expectTaskResult(t, fixture.reporter, "cancel result was not reported")
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("result: %+v", res)
	}
	assertProcessDoesNotStart(t, fixture.running, "provider process should not start after prepare-time cancellation")
}
