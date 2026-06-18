package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

type idlePollSource struct {
	claims chan string
}

func newIdlePollSource() *idlePollSource {
	return &idlePollSource{claims: make(chan string, 16)}
}

func (s *idlePollSource) RegisterRuntime(context.Context, controlplane.RuntimeRegistration) error {
	return nil
}

func (s *idlePollSource) DeregisterRuntime(context.Context, string) error {
	return nil
}

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
