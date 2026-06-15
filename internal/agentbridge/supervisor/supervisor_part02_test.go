package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func (s *runtimeRoutingSource) ClaimTask(_ context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	queue := s.claims[runtimeID]
	if len(queue) == 0 {
		return nil, nil
	}
	req := queue[0]
	s.claims[runtimeID] = queue[1:]
	return &req, nil
}

func (s *runtimeRoutingSource) WatchCancellation(_ context.Context, _ string) (<-chan error, error) {
	return make(chan error), nil
}

type idlePollSource struct {
	claims chan string
}

func (s *idlePollSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}
func (s *idlePollSource) DeregisterRuntime(context.Context, string) error { return nil }
func (s *idlePollSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	return nil
}

func (s *idlePollSource) ClaimTask(_ context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	select {
	case s.claims <- runtimeID:
	default:
	}
	return nil, nil
}

func (s *idlePollSource) WatchCancellation(context.Context, string) (<-chan error, error) {
	return make(chan error), nil
}

func TestSupervisorDefaultMailboxMatchesProviderRuntimeBackpressureSSOT(t *testing.T) {
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID: "daemon-mailbox-default",
		Runtime:  rt,
		Source:   newRuntimeRoutingSource(nil),
		Reporter: newReporterProbe(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := cap(actor.mailbox); got != DefaultMailboxSize {
		t.Fatalf("mailbox size = %d, want %d", got, DefaultMailboxSize)
	}
}

func TestSupervisorBacksOffPollingWhenIdle(t *testing.T) {
	source := &idlePollSource{claims: make(chan string, 16)}
	reporter := newReporterProbe()
	rt := startRuntime(t, process.NewFake())
	actor, err := New(Config{
		DaemonID:       "daemon-1",
		Runtime:        rt,
		Source:         source,
		Reporter:       reporter,
		PollEvery:      10 * time.Millisecond,
		IdlePollEvery:  120 * time.Millisecond,
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
	case <-source.claims:
	case <-time.After(time.Second):
		t.Fatal("first poll did not happen")
	}
	select {
	case runtimeID := <-source.claims:
		t.Fatalf("idle poll happened before backoff elapsed: %s", runtimeID)
	case <-time.After(50 * time.Millisecond):
	}
	select {
	case <-source.claims:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("idle poll did not resume after backoff interval")
	}
}
