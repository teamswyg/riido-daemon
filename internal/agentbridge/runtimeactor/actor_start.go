package runtimeactor

import (
	"context"
	"fmt"
	"time"
)

// Start runs Detect on every adapter, then launches the actor loop.
func (a *Actor) Start(ctx context.Context) error {
	caps, detectedAt, discoveredAt, err := a.detectInitialCapabilities(ctx)
	if err != nil {
		return err
	}
	a.startedAt = discoveredAt
	go a.run(caps, detectedAt)
	close(a.startedCh)
	return nil
}

func (a *Actor) detectInitialCapabilities(ctx context.Context) ([]Capability, map[string]time.Time, time.Time, error) {
	caps := make([]Capability, 0, len(a.cfg.Adapters))
	detectedAt := make(map[string]time.Time, len(a.cfg.Adapters))
	discoveredAt := a.cfg.Now()
	for _, adapter := range a.cfg.Adapters {
		res, err := adapter.Detect(ctx, a.cfg.DetectEnv)
		if err != nil {
			return nil, nil, time.Time{}, fmt.Errorf("runtimeactor: Detect %s: %w", adapter.Name(), err)
		}
		capView, err := buildRuntimeCapability(a.cfg.RuntimeID, adapter.Name(), res, a.cfg.PolicyBundleVersion, discoveredAt)
		if err != nil {
			return nil, nil, time.Time{}, fmt.Errorf("runtimeactor: capability %s: %w", adapter.Name(), err)
		}
		caps = append(caps, capView)
		detectedAt[adapter.Name()] = discoveredAt
	}
	return caps, detectedAt, discoveredAt, nil
}
