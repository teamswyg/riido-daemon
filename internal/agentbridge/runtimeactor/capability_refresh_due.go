package runtimeactor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) refreshDueCapabilities(
	ctx context.Context,
	adapters map[string]agentbridge.Adapter,
	caps []Capability,
	detectedAt map[string]time.Time,
) {
	for idx := range caps {
		capView := caps[idx]
		if !a.capabilityRefreshDue(capView, detectedAt[capView.Provider]) {
			continue
		}
		adapter, ok := adapters[capView.Provider]
		if !ok {
			continue
		}
		refreshed, err := a.detectCapability(ctx, adapter)
		if err != nil {
			continue
		}
		caps[idx] = refreshed
		detectedAt[capView.Provider] = a.cfg.Now()
	}
}
