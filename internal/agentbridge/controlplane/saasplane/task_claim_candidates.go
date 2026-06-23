package saasplane

import (
	"context"
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (p *Plane) claimTaskFromCandidates(ctx context.Context, runtimeID, provider string, candidates []AgentBinding) (*bridge.TaskRequest, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	if len(candidates) == 1 {
		req, err := p.claimTaskFromCandidate(ctx, runtimeID, provider, candidates[0], p.longPollWait(ctx))
		if errors.Is(err, errStaleAgentBindingPoll) {
			return nil, nil
		}
		return req, err
	}
	return p.claimTaskFromCandidatesWithSingleLongPoll(ctx, runtimeID, provider, candidates)
}

func (p *Plane) claimTaskFromCandidatesWithSingleLongPoll(ctx context.Context, runtimeID, provider string, candidates []AgentBinding) (*bridge.TaskRequest, error) {
	survivors := make([]AgentBinding, 0, len(candidates))
	for _, candidate := range candidates {
		req, err := p.claimTaskFromCandidate(ctx, runtimeID, provider, candidate, 0)
		if errors.Is(err, errStaleAgentBindingPoll) {
			continue
		}
		if err != nil || req != nil {
			return req, err
		}
		survivors = append(survivors, candidate)
	}
	if len(survivors) == 0 {
		return nil, nil
	}
	wait := p.longPollWait(ctx)
	if wait <= 0 {
		return nil, nil
	}
	req, err := p.claimTaskFromCandidate(ctx, runtimeID, provider, survivors[0], wait)
	if errors.Is(err, errStaleAgentBindingPoll) {
		return nil, nil
	}
	return req, err
}

func (p *Plane) longPollWait(ctx context.Context) time.Duration {
	if !controlplane.ClaimLongPollEnabled(ctx) {
		return 0
	}
	return p.cfg.LongPollWait
}

func (p *Plane) claimTaskFromCandidate(
	ctx context.Context,
	runtimeID string,
	provider string,
	candidate AgentBinding,
	wait time.Duration,
) (*bridge.TaskRequest, error) {
	poll, err := p.pollAgent(ctx, candidate.AgentID, runtimeID, wait)
	if err != nil {
		return nil, err
	}
	if poll.Assignment == nil {
		return nil, nil
	}
	return p.claimPolledAssignment(ctx, runtimeID, provider, poll)
}
