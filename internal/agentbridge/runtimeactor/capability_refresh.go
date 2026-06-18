package runtimeactor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) capabilityRefreshDue(capView Capability, detectedAt time.Time) bool {
	ttl := a.cfg.CapabilityRefreshEvery
	if ttl < 0 {
		return false
	}
	if !capView.Available && detectedAt.IsZero() {
		return true
	}
	return !detectedAt.IsZero() && a.cfg.Now().Sub(detectedAt) >= ttl
}

func (a *Actor) detectCapability(ctx context.Context, adapter agentbridge.Adapter) (Capability, error) {
	now := a.cfg.Now()
	res, err := adapter.Detect(ctx, a.cfg.DetectEnv)
	if err != nil {
		return Capability{}, err
	}
	return buildRuntimeCapability(a.cfg.RuntimeID, adapter.Name(), res, a.cfg.PolicyBundleVersion, now)
}
