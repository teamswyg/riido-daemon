package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorHeartbeatContinuesDuringWorkspaceMaterialization(t *testing.T) {
	cloneStarted, cloneDone := stubSlowAssignmentClone(t)
	source := &heartbeatDuringPrepareSource{
		req:        slowPrepareTaskRequest(),
		heartbeats: make(chan controlplane.RuntimeHeartbeat, 8),
	}
	reporter := newReporterProbe()
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")

	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		Workdir:        workdir.NewFSAdapter(t.TempDir()),
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: 20 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("supervisor Start: %v", err)
	}
	t.Cleanup(func() {
		stopSupervisorAfterPrepare(t, actor, cloneStarted, cloneDone)
	})

	select {
	case taskID := <-reporter.started:
		if taskID != "asn-slow-prepare" {
			t.Fatalf("started task = %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}
	select {
	case <-cloneStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("worktree materialization did not start")
	}
	drainHeartbeats(source.heartbeats)

	select {
	case hb := <-source.heartbeats:
		if hb.RuntimeID == "" {
			t.Fatalf("heartbeat missing runtime id: %+v", hb)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("heartbeat was blocked by workspace materialization")
	}
}
