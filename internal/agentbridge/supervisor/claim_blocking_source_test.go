package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
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

func (s *blockingClaimSource) DeregisterRuntime(context.Context, string) error { return nil }

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
