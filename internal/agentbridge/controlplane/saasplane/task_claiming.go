package saasplane

import (
	"context"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func (p *Plane) ClaimTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	provider := providerFromRuntimeID(runtimeID)
	if p.dynamicBindingsEnabled() {
		return p.claimDynamicBindingTask(ctx, runtimeID, provider)
	}
	return p.claimStaticBindingTask(ctx, runtimeID, provider)
}

func (p *Plane) claimDynamicBindingTask(ctx context.Context, runtimeID, provider string) (*bridge.TaskRequest, error) {
	bindings, err := p.agentBindings(ctx)
	if err != nil {
		return nil, err
	}
	candidates := dynamicBindingCandidates(runtimeID, provider, bindings, "")
	refresh := func(ctx context.Context, stale AgentBinding) ([]AgentBinding, error) {
		bindings, err := p.agentBindings(ctx)
		if err != nil {
			return nil, err
		}
		p.excludeAgentBindingFromCache(ctx, stale.AgentID)
		return dynamicBindingCandidates(runtimeID, provider, bindings, stale.AgentID), nil
	}
	return p.claimTaskFromCandidatesWithRefresh(ctx, runtimeID, provider, candidates, refresh)
}

func dynamicBindingCandidates(
	runtimeID string,
	provider string,
	bindings []assignmentcontract.AgentRuntimeBinding,
	excludedAgentID string,
) []AgentBinding {
	candidates := make([]AgentBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.RuntimeProvider != provider || strings.TrimSpace(binding.RuntimeID) != strings.TrimSpace(runtimeID) {
			continue
		}
		if binding.AgentID == excludedAgentID {
			continue
		}
		candidates = append(candidates, AgentBinding{
			AgentID:         binding.AgentID,
			RuntimeProvider: binding.RuntimeProvider,
		})
	}
	return candidates
}

func (p *Plane) claimStaticBindingTask(ctx context.Context, runtimeID, provider string) (*bridge.TaskRequest, error) {
	runtimeAgent, hasRuntimeAgent := agentFromRuntimeID(runtimeID)
	candidates := make([]AgentBinding, 0, len(p.cfg.Agents))
	for _, agent := range p.cfg.Agents {
		if agent.RuntimeProvider != provider {
			continue
		}
		if hasRuntimeAgent && agent.AgentID != runtimeAgent {
			continue
		}
		candidates = append(candidates, agent)
	}
	return p.claimTaskFromCandidates(ctx, runtimeID, provider, candidates)
}
