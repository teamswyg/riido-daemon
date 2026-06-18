package saasplane

import (
	"context"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (p *Plane) agentBindings(ctx context.Context) ([]assignmentcontract.AgentRuntimeBinding, error) {
	now := time.Now()
	cached, ok, err := p.cachedAgentBindings(ctx, now)
	if err != nil || ok {
		return cached, err
	}
	var out AgentRuntimeBindingListResponse
	if err := p.getJSON(ctx, "/v1/daemon/agent-bindings", &out); err != nil {
		return nil, err
	}
	bindings := cloneAgentRuntimeBindings(out.Bindings)
	_ = p.withState(ctx, func(s *planeState) {
		s.agentBindingsCache = cloneAgentRuntimeBindings(bindings)
		s.agentBindingsCachedAt = now
	})
	return bindings, nil
}

func (p *Plane) cachedAgentBindings(ctx context.Context, now time.Time) ([]assignmentcontract.AgentRuntimeBinding, bool, error) {
	var bindings []assignmentcontract.AgentRuntimeBinding
	var ok bool
	err := p.withState(ctx, func(s *planeState) {
		if s.agentBindingsCachedAt.IsZero() || now.Sub(s.agentBindingsCachedAt) >= agentBindingCacheTTL {
			return
		}
		bindings = cloneAgentRuntimeBindings(s.agentBindingsCache)
		ok = true
	})
	return bindings, ok, err
}

func (p *Plane) invalidateAgentBindingsCache(ctx context.Context) {
	_ = p.withState(ctx, func(s *planeState) {
		s.agentBindingsCache = nil
		s.agentBindingsCachedAt = time.Time{}
	})
}

func cloneAgentRuntimeBindings(in []assignmentcontract.AgentRuntimeBinding) []assignmentcontract.AgentRuntimeBinding {
	if len(in) == 0 {
		return nil
	}
	return append([]assignmentcontract.AgentRuntimeBinding(nil), in...)
}
