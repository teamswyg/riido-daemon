package supervisor

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorUsesLogicalTaskIDMetadataForWorkspace(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "asn-1",
		Provider: "fake",
		Prompt:   "hello",
		Metadata: map[string]string{
			MetadataWorkspaceID:         "ws-1",
			MetadataRunID:               "asn-1",
			controlplane.MetadataTaskID: "task-a",
		},
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
		Workdir:        workdir.NewFSAdapter(t.TempDir()),
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
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
		if taskID != "asn-1" {
			t.Fatalf("started execution: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	go func() {
		running.EmitStdout([]byte("done"))
		running.EmitExit(0, nil)
	}()

	select {
	case res := <-reporter.results:
		wantSuffix := filepath.Join("ws-1", "tasks", "task-a", "runs", "asn-1", "workdir")
		if !strings.HasSuffix(res.Workdir, wantSuffix) {
			t.Fatalf("workdir = %q, want suffix %q", res.Workdir, wantSuffix)
		}
		events := readRunEvents(t, filepath.Join(filepath.Dir(res.Workdir), "ir", "events.jsonl"))
		assertRunEvent(t, events, ir.EventRunReportedDone, func(ev ir.CanonicalEvent) {
			if ev.TaskID != "task-a" || ev.RunID != "asn-1" {
				t.Fatalf("logical task/run ids not preserved: %+v", ev)
			}
		})
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}
}

func TestSupervisorDispatchesTaskToSelectedRuntimeActor(t *testing.T) {
	source := newRuntimeRoutingSource(map[string][]bridge.TaskRequest{
		"rt-codex": {{
			ID:                       "t-codex",
			Provider:                 "codex",
			Prompt:                   "hello",
			AllowExperimentalRuntime: true,
			Metadata: map[string]string{
				MetadataWorkspaceID: "ws-1",
			},
		}},
	})
	reporter := newReporterProbe()
	claudeFake := process.NewFake()
	codexFake := process.NewFake()
	codexRunning := process.NewFakeRunning()
	codexFake.NextRunning = codexRunning
	rtClaude := startNamedRuntime(t, claudeFake, "rt-claude", "claude")
	rtCodex := startNamedRuntime(t, codexFake, "rt-codex", "codex")

	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtimes:       []*runtimeactor.Actor{rtClaude, rtCodex},
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
		t.Fatalf("supervisor Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	seenRegistrations := map[string]bool{}
	for range 2 {
		select {
		case rt := <-source.registered:
			seenRegistrations[rt.RuntimeID] = true
			if rt.Provider != strings.TrimPrefix(rt.RuntimeID, "rt-") {
				t.Fatalf("provider-specific registration mismatch: %+v", rt)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("runtime registration was not published")
		}
	}
	if !seenRegistrations["rt-claude"] || !seenRegistrations["rt-codex"] {
		t.Fatalf("runtime registrations missing: %+v", seenRegistrations)
	}

	select {
	case taskID := <-reporter.started:
		if taskID != "t-codex" {
			t.Fatalf("started task: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not dispatched")
	}
	select {
	case cmd := <-codexRunning.StartedRecv():
		if cmd.Executable != "codex" {
			t.Fatalf("codex runtime command mismatch: %+v", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("codex runtime did not spawn process")
	}

	go func() {
		codexRunning.EmitStdout([]byte("done"))
		codexRunning.EmitExit(0, nil)
	}()
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("result was not reported")
	}
}

func hasEnvPrefix(env []string, prefix string) bool {
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return true
		}
	}
	return false
}
