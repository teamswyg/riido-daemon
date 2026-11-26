package supervisor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type blockedRuntimeHeartbeatSource struct {
	*controlplane.MemorySource
	slowStarted chan struct{}
	releaseSlow chan struct{}
	healthy     chan struct{}
	slowCalls   atomic.Int32
}

func (s *blockedRuntimeHeartbeatSource) Heartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	if hb.RuntimeID == "rt-slow" {
		if s.slowCalls.Add(1) == 1 {
			close(s.slowStarted)
		}
		// Hold this request past its deadline to simulate a transport still unwinding.
		<-s.releaseSlow
		return ctx.Err()
	}
	select {
	case s.healthy <- struct{}{}:
	case <-ctx.Done():
	}
	return nil
}

func TestHeartbeatContinuesWhileAnotherRuntimeIsHeld(t *testing.T) {
	source := &blockedRuntimeHeartbeatSource{
		MemorySource: controlplane.NewMemorySource(),
		slowStarted:  make(chan struct{}),
		releaseSlow:  make(chan struct{}),
		healthy:      make(chan struct{}, 4),
	}
	runtimes := []*runtimeactor.Actor{
		startNamedRuntime(t, process.NewFake(), "rt-slow", "claude"),
		startNamedRuntime(t, process.NewFake(), "rt-healthy", "codex"),
	}
	actor := &Actor{cfg: Config{Source: source, HeartbeatEvery: 20 * time.Millisecond}}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		actor.heartbeatLoop(ctx, runtimes)
	}()
	t.Cleanup(func() {
		cancel()
		close(source.releaseSlow)
		<-done
	})

	timeout := time.After(time.Second)
	select {
	case <-source.slowStarted:
	case <-timeout:
		t.Fatal("slow runtime heartbeat did not start")
	}
	for range 2 {
		select {
		case <-source.healthy:
		case <-timeout:
			t.Fatal("healthy runtime heartbeat stalled behind the slow runtime")
		}
	}
	if calls := source.slowCalls.Load(); calls != 1 {
		t.Fatalf("slow runtime heartbeat started %d times while the first request was held", calls)
	}
}

type blockedClaimHeartbeatSource struct {
	*controlplane.MemorySource
	entered    chan struct{}
	heartbeats chan struct{}
}

func (s *blockedClaimHeartbeatSource) ClaimTask(ctx context.Context, _ string) (*bridge.TaskRequest, error) {
	select {
	case s.entered <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return nil, ctx.Err()
}

func (s *blockedClaimHeartbeatSource) Heartbeat(context.Context, controlplane.RuntimeHeartbeat) error {
	select {
	case s.heartbeats <- struct{}{}:
	default:
	}
	return nil
}

func TestHeartbeatContinuesWhileClaimIsHeld(t *testing.T) {
	source := &blockedClaimHeartbeatSource{MemorySource: controlplane.NewMemorySource(), entered: make(chan struct{}, 1), heartbeats: make(chan struct{}, 4)}
	actor, err := New(Config{DaemonID: "heartbeat", RiidoDaemonVersion: "test", Runtime: startRuntime(t, process.NewFake()), Source: source, Reporter: newReporterProbe(), PollEvery: time.Millisecond, HeartbeatEvery: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if err = actor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	}()
	select {
	case <-source.entered:
	case <-time.After(time.Second):
		t.Fatal("claim never held")
	}
	for range 3 {
		select {
		case <-source.heartbeats:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("heartbeat stalled behind claim")
		}
	}
}
