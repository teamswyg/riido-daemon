package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

const supervisorCancellationTestTimeout = 5 * time.Second

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
