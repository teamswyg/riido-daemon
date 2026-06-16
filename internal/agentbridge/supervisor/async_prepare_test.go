package supervisor

import (
	"context"
	"sync"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type heartbeatDuringPrepareSource struct {
	mu         sync.Mutex
	req        bridge.TaskRequest
	claimed    bool
	heartbeats chan controlplane.RuntimeHeartbeat
}

func (s *heartbeatDuringPrepareSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}

func (s *heartbeatDuringPrepareSource) DeregisterRuntime(context.Context, string) error {
	return nil
}

func (s *heartbeatDuringPrepareSource) Heartbeat(_ context.Context, hb controlplane.RuntimeHeartbeat) error {
	select {
	case s.heartbeats <- hb:
	default:
	}
	return nil
}

func (s *heartbeatDuringPrepareSource) ClaimTask(context.Context, string) (*bridge.TaskRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.claimed {
		return nil, nil
	}
	s.claimed = true
	req := s.req
	return &req, nil
}

func (s *heartbeatDuringPrepareSource) WatchCancellation(context.Context, string) (<-chan error, error) {
	return make(chan error), nil
}

func TestSupervisorHeartbeatContinuesDuringWorkspaceMaterialization(t *testing.T) {
	cloneStarted := make(chan struct{})
	cloneDone := make(chan struct{})
	unblockClone := make(chan struct{})
	var unblockOnce sync.Once

	originalGitResolver := resolveAssignmentGitExecutable
	originalClone := runAssignmentGitClone
	resolveAssignmentGitExecutable = func(string, string) (string, bool) {
		return "/usr/bin/git", true
	}
	runAssignmentGitClone = func(ctx context.Context, _ string, _ []string) error {
		close(cloneStarted)
		defer close(cloneDone)
		select {
		case <-unblockClone:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	t.Cleanup(func() {
		unblockOnce.Do(func() { close(unblockClone) })
		runAssignmentGitClone = originalClone
		resolveAssignmentGitExecutable = originalGitResolver
	})

	source := &heartbeatDuringPrepareSource{
		req: bridge.TaskRequest{
			ID:       "asn-slow-prepare",
			Provider: "fake",
			Prompt:   "slow prepare",
			Worktree: &assignmentcontract.AssignmentWorktree{
				RepositoryFullName: "teamswyg/riido-daemon",
				RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
				BranchName:         "RIID-4964-agent-profile-upload",
			},
			Metadata: map[string]string{
				MetadataWorkspaceID:         "ws-1",
				MetadataRunID:               "asn-slow-prepare",
				controlplane.MetadataTaskID: "task-slow-prepare",
			},
		},
		heartbeats: make(chan controlplane.RuntimeHeartbeat, 8),
	}
	reporter := newReporterProbe()
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	rt := startRuntime(t, fake)

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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
		select {
		case <-cloneStarted:
			select {
			case <-cloneDone:
			case <-time.After(time.Second):
				t.Error("workspace materialization goroutine did not stop")
			}
		default:
		}
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

func drainHeartbeats(ch <-chan controlplane.RuntimeHeartbeat) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
