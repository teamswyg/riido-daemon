package runtimeactor

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) capabilityForSubmit(
	ctx context.Context,
	adapter agentbridge.Adapter,
	caps []Capability,
	detectedAt map[string]time.Time,
	provider string,
) (Capability, bool, error) {
	idx := capabilityIndexForProvider(caps, provider)
	if idx < 0 {
		return Capability{}, false, nil
	}
	capView := caps[idx]
	if !a.capabilityRefreshDue(capView, detectedAt[provider]) {
		return capView, true, nil
	}
	refreshed, err := a.detectCapability(ctx, adapter)
	if err != nil {
		if capView.Available {
			return capView, true, nil
		}
		return Capability{}, true, fmt.Errorf("runtimeactor: Detect %s: %w", adapter.Name(), err)
	}
	caps[idx] = refreshed
	detectedAt[provider] = a.cfg.Now()
	return refreshed, true, nil
}
