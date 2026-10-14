package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type taskResultSupervisorRun struct {
	source   *controlplane.MemorySource
	reporter *reporterProbe
	running  *process.FakeRunning
}

func startTaskResultSupervisor(t *testing.T) taskResultSupervisorRun {
	t.Helper()
	source := controlplane.NewMemorySource()
	source.Enqueue(taskResultRequest())
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)
	actor, err := New(Config{
		DaemonID:           "daemon-1",
		RiidoDaemonVersion: "riido-agentd v1.2.3",
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
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})
	return taskResultSupervisorRun{source: source, reporter: reporter, running: running}
}
