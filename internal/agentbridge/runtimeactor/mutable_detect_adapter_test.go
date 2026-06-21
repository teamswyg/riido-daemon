package runtimeactor

import (
	"context"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type mutableDetectAdapter struct {
	stubAdapter
	mu sync.Mutex
}

func (a *mutableDetectAdapter) Detect(
	_ context.Context,
	_ agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.detected, nil
}

func (a *mutableDetectAdapter) setDetected(res agentbridge.DetectResult) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.detected = res
}
