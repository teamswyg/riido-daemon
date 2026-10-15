package runtimeactor

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (r *capabilityRefresh) scheduleDue(a *Actor, adapters map[string]agentbridge.Adapter, caps []Capability, detectedAt map[string]time.Time) {
	for _, capView := range caps {
		if a.capabilityRefreshDue(capView, detectedAt[capView.Provider]) {
			if adapter := adapters[capView.Provider]; adapter != nil {
				r.start(a, adapter, capView)
			}
		}
	}
}
