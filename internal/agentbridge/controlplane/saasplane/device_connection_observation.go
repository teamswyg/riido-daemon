package saasplane

import (
	"context"
	"log"
)

func (p *Plane) observeDeviceConnections(ctx context.Context, observation AgentRuntimeBindingListResponse) {
	if observation.ConnectionRevision == "" {
		return
	}
	var previousRevision string
	var previousPrincipalCount int
	_ = p.withState(ctx, func(s *planeState) {
		previousRevision = s.connectionRevision
		previousPrincipalCount = s.connectedPrincipalCount
		s.connectionRevision = observation.ConnectionRevision
		s.connectedPrincipalCount = observation.ConnectedPrincipalCount
	})
	if previousRevision == observation.ConnectionRevision && previousPrincipalCount == observation.ConnectedPrincipalCount {
		return
	}
	log.Printf(
		"riido_daemon event=device_connections_observed previous_principal_count=%d connected_principal_count=%d connection_revision=%q bindings=%d",
		previousPrincipalCount,
		observation.ConnectedPrincipalCount,
		observation.ConnectionRevision,
		len(observation.Bindings),
	)
}
