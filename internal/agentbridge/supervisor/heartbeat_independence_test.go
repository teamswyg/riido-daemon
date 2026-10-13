package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

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
