package supervisor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestSupervisorStopLifecyclePropagatesForcedLevel(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "t-forced-stop",
		Provider: "fake",
		Prompt:   "x",
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-forced-stop",
			MetadataRunID:       "run-forced-stop",
		},
	})

	reporter := newLifecycleReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)

	actor, err := New(Config{
		DaemonID:           "daemon-forced-stop",
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
		t.Fatal(err)
	}

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}
	select {
	case <-running.StartedRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not spawned")
	}

	shutdownCtx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownForced, 2*time.Second)
	defer cancel()
	if err := actor.StopLifecycle(shutdownCtx); err != nil {
		t.Fatalf("StopLifecycle: %v", err)
	}

	select {
	case level := <-reporter.completeLevels:
		if level != lifecycle.ShutdownForced {
			t.Fatalf("CompleteTask lifecycle level = %s, want %s", level, lifecycle.ShutdownForced)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown result was not reported")
	}
	select {
	case <-running.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not killed on forced supervisor stop")
	}
}

func TestSupervisorWorkdirRequiresWorkspaceID(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{ID: "t-no-workspace", Provider: "fake", Prompt: "x"})

	reporter := newReporterProbe()
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		Workdir:        workdir.NewFSAdapter(t.TempDir()),
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultFailed || res.Error == "" {
			t.Fatalf("expected workdir failure result, got %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("workdir failure was not reported")
	}
}

func TestSupervisorBlocksTaskWhenRequiredSurfaceUnsupported(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:               "t-needs-mcp",
		Provider:         "fake",
		Prompt:           "x",
		RequiredSurfaces: []string{"mcp"},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)
	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case <-reporter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultBlocked {
			t.Fatalf("expected blocked result, got %+v", res)
		}
		if !strings.Contains(res.Error, "MISSING_REQUIRED_SURFACE:mcp") {
			t.Fatalf("missing scheduler reason: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ineligible task was not reported")
	}
	if running.Command().Executable != "" {
		t.Fatalf("provider process should not have spawned: %+v", running.Command())
	}
}
