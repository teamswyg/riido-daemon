package saasplane

import (
	"context"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func (p *Plane) claimTaskFromCandidates(ctx context.Context, runtimeID, provider string, candidates []AgentBinding) (*bridge.TaskRequest, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	if len(candidates) == 1 {
		return p.claimTaskFromCandidate(ctx, runtimeID, provider, candidates[0], p.cfg.LongPollWait)
	}
	return p.claimTaskFromCandidatesWithSingleLongPoll(ctx, runtimeID, provider, candidates)
}

func (p *Plane) claimTaskFromCandidatesWithSingleLongPoll(ctx context.Context, runtimeID, provider string, candidates []AgentBinding) (*bridge.TaskRequest, error) {
	for _, candidate := range candidates {
		req, err := p.claimTaskFromCandidate(ctx, runtimeID, provider, candidate, 0)
		if err != nil || req != nil {
			return req, err
		}
	}
	return p.claimTaskFromCandidate(ctx, runtimeID, provider, candidates[0], p.cfg.LongPollWait)
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

func (p *Plane) claimPolledAssignment(
	ctx context.Context,
	runtimeID string,
	provider string,
	poll assignmentcontract.PollResponse,
) (*bridge.TaskRequest, error) {
	switch poll.Action {
	case assignmentcontract.PollStart, assignmentcontract.PollActive:
		return p.claimStartOrActiveAssignment(ctx, runtimeID, provider, poll)
	case assignmentcontract.PollCancel:
		_ = p.deliverCancel(ctx, *poll.Assignment)
		return nil, nil
	default:
		return nil, nil
	}
}
