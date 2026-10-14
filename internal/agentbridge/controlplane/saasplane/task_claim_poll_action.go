package saasplane

import (
	"context"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

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
