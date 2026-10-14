package saasplane

import (
	"context"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func (p *Plane) claimStartOrActiveAssignment(
	ctx context.Context,
	runtimeID string,
	provider string,
	poll assignmentcontract.PollResponse,
) (*bridge.TaskRequest, error) {
	assignment := *poll.Assignment
	if assignment.RuntimeProvider != "" && assignment.RuntimeProvider != provider {
		return nil, nil
	}
	if poll.Action == assignmentcontract.PollActive && assignmentResumeSessionID(assignment) == "" {
		return nil, nil
	}
	if err := p.saveAssignmentRuntime(ctx, assignment, runtimeID); err != nil {
		return nil, err
	}
	return taskRequestFromAssignment(assignment), nil
}
