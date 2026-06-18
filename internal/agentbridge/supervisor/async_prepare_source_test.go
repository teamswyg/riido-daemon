package supervisor

import (
	"context"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
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

func drainHeartbeats(ch <-chan controlplane.RuntimeHeartbeat) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
