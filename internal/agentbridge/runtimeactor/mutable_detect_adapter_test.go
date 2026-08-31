package runtimeactor

import (
	"context"
	"errors"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type mutableDetectAdapter struct {
	stubAdapter
	mu        sync.Mutex
	detectErr error
}

func (a *mutableDetectAdapter) Detect(
	_ context.Context,
	_ agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.detected, a.detectErr
}

func (a *mutableDetectAdapter) failDetection() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.detectErr = errors.New("sensitive provider failure")
}

func (a *mutableDetectAdapter) setDetected(res agentbridge.DetectResult) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.detected = res
}
