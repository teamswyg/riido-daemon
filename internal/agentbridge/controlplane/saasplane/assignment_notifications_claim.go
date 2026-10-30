package saasplane

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (p *Plane) claimNotifiedTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, bool, error) {
	p.notices.mu.Lock()
	enabled := p.notices.started && !p.notices.legacy
	bindings := cloneAgentRuntimeBindings(p.notices.bindings)
	p.notices.mu.Unlock()
	if !enabled {
		return nil, false, nil
	}
	if !controlplane.ClaimCapacityAvailable(ctx) {
		return nil, true, nil
	}
	provider := providerFromRuntimeID(runtimeID)
	for _, candidate := range dynamicBindingCandidates(runtimeID, provider, bindings, "") {
		p.notices.mu.Lock()
		notice, pending := p.notices.pending[candidate.AgentID]
		p.notices.mu.Unlock()
		if !pending || time.Now().Before(notice.retryAfter) {
			continue
		}
		// Never hold the supervisor for a long poll when the stream has already woken it.
		request, err := p.claimTaskFromCandidate(ctx, runtimeID, provider, candidate, 0)
		p.notices.mu.Lock()
		current, exists := p.notices.pending[candidate.AgentID]
		if exists && current.generation == notice.generation {
			if err == nil {
				delete(p.notices.pending, candidate.AgentID)
			} else {
				current.retryAfter = time.Now().Add(5 * time.Second)
				p.notices.pending[candidate.AgentID] = current
			}
		}
		p.notices.mu.Unlock()
		if err != nil || request != nil {
			return request, true, err
		}
	}
	return nil, true, nil
}
