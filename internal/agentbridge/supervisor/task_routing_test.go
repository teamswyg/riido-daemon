package supervisor

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorUsesLogicalTaskIDMetadataForWorkspace(t *testing.T) {
	var gotCloneArgs []string
	originalClone := runAssignmentGitClone
	runAssignmentGitClone = func(_ context.Context, _ string, args []string) error {
		gotCloneArgs = append([]string(nil), args...)
		return nil
	}
	t.Cleanup(func() { runAssignmentGitClone = originalClone })

	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:                       "asn-1",
		Provider:                 "codex",
		Prompt:                   "hello",
		AllowExperimentalRuntime: true,
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
			BranchName:         "RIID-4964-agent-profile-upload",
		},
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
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")

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
		if len(gotCloneArgs) == 0 || gotCloneArgs[len(gotCloneArgs)-1] != res.Workdir {
			t.Fatalf("clone args did not target prepared workdir: args=%#v workdir=%q", gotCloneArgs, res.Workdir)
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

func TestSupervisorBlocksPrivateAssignmentWorktreeBeforeProviderStart(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:                       "asn-private",
		Provider:                 "codex",
		Prompt:                   "fix private repo",
		AllowExperimentalRuntime: true,
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/private-repo",
			RepositoryURL:      "https://github.com/teamswyg/private-repo",
			BranchName:         "main",
			IsPrivate:          true,
		},
		Metadata: map[string]string{
			MetadataWorkspaceID:         "ws-1",
			MetadataRunID:               "asn-private",
			controlplane.MetadataTaskID: "task-private",
		},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")

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
		if taskID != "asn-private" {
			t.Fatalf("started execution: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultBlocked {
			t.Fatalf("result status = %s, want %s; result=%+v", res.Status, agentbridge.ResultBlocked, res)
		}
		if !strings.Contains(res.Error, "private assignment worktree requires git credentials") {
			t.Fatalf("result error does not explain private worktree credentials: %q", res.Error)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("blocked result was not reported")
	}
	select {
	case <-running.StartedRecv():
		t.Fatal("provider process should not start for blocked private worktree")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestSupervisorBlocksWorktreeWhenRuntimeSurfaceUnsupported(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "asn-worktree-unsupported",
		Provider: "fake",
		Prompt:   "fix repo",
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
			BranchName:         "main",
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
		if taskID != "asn-worktree-unsupported" {
			t.Fatalf("started execution: %q", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("task was not claimed")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultBlocked {
			t.Fatalf("result status = %s, want %s; result=%+v", res.Status, agentbridge.ResultBlocked, res)
		}
		if !strings.Contains(res.Error, "MISSING_REQUIRED_SURFACE:worktree") {
			t.Fatalf("result error should explain missing worktree surface: %q", res.Error)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("blocked result was not reported")
	}
	select {
	case <-running.StartedRecv():
		t.Fatal("provider process should not start when worktree surface is unsupported")
	case <-time.After(100 * time.Millisecond):
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
