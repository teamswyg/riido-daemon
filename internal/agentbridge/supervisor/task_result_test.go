package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorClaimsTaskAndReportsResult(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "t-1",
		Provider: "fake",
		Prompt:   "hello",
		Metadata: map[string]string{
			MetadataWorkspaceID:                   "ws-1",
			MetadataAgentName:                     "Riido",
			agentbridge.MetadataTelemetryContract: agentbridge.TelemetryPlacementPrompt,
		},
	})

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

	select {
	case taskID := <-reporter.started:
		if taskID != "t-1" {
			t.Fatalf("started task: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventLifecycle || ev.Phase != agentbridge.StateRunning {
			t.Fatalf("running event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("running event was not reported")
	}

	go func() {
		running.EmitStdout([]byte("done"))
		running.EmitExit(0, nil)
	}()

	select {
	case ev := <-reporter.events:
		if ev.Kind != agentbridge.EventTextDelta || ev.Text != "done" {
			t.Fatalf("event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("event was not reported")
	}

	select {
	case res := <-reporter.results:
		assertSupervisorCompletedRun(t, res, running)
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}

	registered := source.Registered()
	if len(registered) != 1 || registered[0].RuntimeID != "rt-local" || registered[0].DaemonID != "daemon-1" {
		t.Fatalf("registered runtime: %+v", registered)
	}
	if _, ok := registered[0].Capabilities["provider.fake.supports_mcp"]; !ok {
		t.Fatalf("registered runtime missing capability support flags: %+v", registered[0].Capabilities)
	}
	if _, ok := registered[0].Capabilities["provider.fake.supports_worktree"]; !ok {
		t.Fatalf("registered runtime missing worktree support flag: %+v", registered[0].Capabilities)
	}
	if registered[0].CapabilityAttributes["provider.fake.provider_version"] == "" {
		t.Fatalf("registered runtime missing provider version attribute: %+v", registered[0].CapabilityAttributes)
	}
}
