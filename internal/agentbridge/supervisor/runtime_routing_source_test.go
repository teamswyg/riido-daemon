package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

type runtimeRoutingSource struct {
	claims       map[string][]bridge.TaskRequest
	registered   chan controlplane.RuntimeRegistration
	deregistered chan string
}

func newRuntimeRoutingSource(claims map[string][]bridge.TaskRequest) *runtimeRoutingSource {
	return &runtimeRoutingSource{
		claims:       claims,
		registered:   make(chan controlplane.RuntimeRegistration, 8),
		deregistered: make(chan string, 8),
	}
}

func (s *runtimeRoutingSource) RegisterRuntime(_ context.Context, rt controlplane.RuntimeRegistration) error {
	s.registered <- rt
	return nil
}

func (s *runtimeRoutingSource) DeregisterRuntime(_ context.Context, runtimeID string) error {
	s.deregistered <- runtimeID
	return nil
}

func (s *runtimeRoutingSource) Heartbeat(_ context.Context, _ controlplane.RuntimeHeartbeat) error {
	return nil
}
