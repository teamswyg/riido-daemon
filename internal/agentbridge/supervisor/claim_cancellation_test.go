package supervisor

import (
	"context"
	"testing"
	"time"

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
