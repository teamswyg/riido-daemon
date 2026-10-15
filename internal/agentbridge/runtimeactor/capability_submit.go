package runtimeactor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Freshness is resolved asynchronously before handleSubmit enters the mailbox path.
func (a *Actor) capabilityForSubmit(_ context.Context, _ agentbridge.Adapter, caps []Capability, _ map[string]time.Time, provider string) (Capability, bool, error) {
	idx := capabilityIndexForProvider(caps, provider)
	if idx < 0 {
		return Capability{}, false, nil
	}
	return caps[idx], true, nil
}
