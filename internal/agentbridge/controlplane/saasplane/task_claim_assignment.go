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
	scope, err := p.saveClaimRuntime(ctx, assignment, runtimeID, poll.Action == assignmentcontract.PollStart)
	if err != nil {
		return nil, err
	}
	req := taskRequestFromAssignment(assignment)
	req.Metadata["execution_device_id"] = p.cfg.DeviceID
	req.LocalExecutionScope = scope
	return req, nil
}
