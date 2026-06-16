package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type cancelSource struct {
	req       bridge.TaskRequest
	cancel    chan error
	watchCtxs chan context.Context
}

func (s *cancelSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}
func (s *cancelSource) DeregisterRuntime(context.Context, string) error { return nil }
func (s *cancelSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	return nil
}

func (s *cancelSource) ClaimTask(context.Context, string) (*bridge.TaskRequest, error) {
	if s.req.ID == "" {
		return nil, nil
	}
	req := s.req
	s.req = bridge.TaskRequest{}
	return &req, nil
}

func (s *cancelSource) WatchCancellation(ctx context.Context, _ string) (<-chan error, error) {
	if s.watchCtxs != nil {
		select {
		case s.watchCtxs <- ctx:
		default:
		}
	}
	return s.cancel, nil
}

func TestSupervisorRoutesCancellationToRuntime(t *testing.T) {
	source := &cancelSource{
		req:    bridge.TaskRequest{ID: "t-cancel", Provider: "fake", Prompt: "x"},
		cancel: make(chan error, 1),
	}
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
		HeartbeatEvery: time.Second,
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

	source.cancel <- errors.New("human cancel")

	select {
	case <-running.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not killed")
	}

	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancel result was not reported")
	}
}

func TestSupervisorCancelsCancellationWatcherOnComplete(t *testing.T) {
	source := &cancelSource{
		req:       bridge.TaskRequest{ID: "t-complete", Provider: "fake", Prompt: "x"},
		cancel:    make(chan error),
		watchCtxs: make(chan context.Context, 1),
	}
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
		HeartbeatEvery: time.Second,
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
	case <-running.StartedRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("provider process was not started")
	}
	var watchCtx context.Context
	select {
	case watchCtx = <-source.watchCtxs:
	case <-time.After(2 * time.Second):
		t.Fatal("cancellation watcher was not started")
	}

	running.EmitStdout([]byte("done"))
	running.EmitExit(0, nil)
	select {
	case res := <-reporter.results:
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("completion was not reported")
	}
	select {
	case <-watchCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("cancellation watcher context was not cancelled after completion")
	}
}
