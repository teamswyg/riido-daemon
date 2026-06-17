package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type blockingClaimSource struct {
	started  chan struct{}
	canceled chan struct{}
}

func newBlockingClaimSource() *blockingClaimSource {
	return &blockingClaimSource{
		started:  make(chan struct{}),
		canceled: make(chan struct{}),
	}
}

func (s *blockingClaimSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}

func (s *blockingClaimSource) DeregisterRuntime(context.Context, string) error {
	return nil
}

func (s *blockingClaimSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	return nil
}

func (s *blockingClaimSource) ClaimTask(ctx context.Context, _ string) (*bridge.TaskRequest, error) {
	select {
	case <-s.started:
	default:
		close(s.started)
	}
	<-ctx.Done()
	close(s.canceled)
	return nil, ctx.Err()
}

func (s *blockingClaimSource) WatchCancellation(context.Context, string) (<-chan error, error) {
	return make(chan error), nil
}

func TestSupervisorStopCancelsInFlightClaim(t *testing.T) {
	source := newBlockingClaimSource()
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID:       "daemon-claim-cancel",
		Runtime:        rt,
		Source:         source,
		Reporter:       newReporterProbe(),
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-source.started:
	case <-time.After(time.Second):
		t.Fatal("claim did not start")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if err := actor.Stop(stopCtx); err != nil {
		t.Fatalf("Stop should cancel in-flight claim: %v", err)
	}
	select {
	case <-source.canceled:
	default:
		t.Fatal("claim context was not canceled")
	}
}

type watchCleanupSource struct {
	task           bridge.TaskRequest
	claimed        bool
	watchStarted   chan struct{}
	watchCtxClosed chan struct{}
}

func newWatchCleanupSource() *watchCleanupSource {
	return &watchCleanupSource{
		task: bridge.TaskRequest{
			ID:       "t-watch-cleanup",
			Provider: "fake",
			Prompt:   "hello",
		},
		watchStarted:   make(chan struct{}),
		watchCtxClosed: make(chan struct{}),
	}
}

func (s *watchCleanupSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}

func (s *watchCleanupSource) DeregisterRuntime(context.Context, string) error {
	return nil
}

func (s *watchCleanupSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	return nil
}

func (s *watchCleanupSource) ClaimTask(context.Context, string) (*bridge.TaskRequest, error) {
	if s.claimed {
		return nil, nil
	}
	s.claimed = true
	task := s.task
	return &task, nil
}

func (s *watchCleanupSource) WatchCancellation(ctx context.Context, _ string) (<-chan error, error) {
	ch := make(chan error)
	close(s.watchStarted)
	go func() {
		<-ctx.Done()
		close(s.watchCtxClosed)
	}()
	return ch, nil
}

func TestSupervisorCancelsCancellationWatchAfterTaskCompletion(t *testing.T) {
	source := newWatchCleanupSource()
	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startRuntime(t, fake)

	actor, err := New(Config{
		DaemonID:       "daemon-watch-cleanup",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		PollEvery:      10 * time.Millisecond,
		HeartbeatEvery: time.Hour,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})

	select {
	case taskID := <-reporter.started:
		if taskID != source.task.ID {
			t.Fatalf("started task = %q, want %q", taskID, source.task.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("task did not start")
	}
	select {
	case <-source.watchStarted:
	case <-time.After(time.Second):
		t.Fatal("cancellation watch did not start")
	}
	select {
	case <-running.StartedRecv():
	case <-time.After(time.Second):
		t.Fatal("provider process did not start")
	}

	running.EmitStdout([]byte("done"))
	running.EmitExit(0, nil)
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result status = %s, want %s; result=%+v", res.Status, agentbridge.ResultCompleted, res)
		}
	case <-time.After(time.Second):
		t.Fatal("task result was not reported")
	}
	select {
	case <-source.watchCtxClosed:
	case <-time.After(time.Second):
		t.Fatal("cancellation watch context was not canceled after completion")
	}
}
