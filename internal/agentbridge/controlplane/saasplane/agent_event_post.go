package saasplane

import (
	"context"
	"net/url"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (p *Plane) postAgentEvent(
	ctx context.Context,
	assignment assignmentcontract.Assignment,
	req assignmentcontract.AgentEventRequest,
) (assignmentcontract.AgentEventResponse, error) {
	var out assignmentcontract.AgentEventResponse
	req.DaemonID = p.cfg.DaemonID
	req.DeviceID = p.cfg.DeviceID
	runtimeID, err := p.runtimeIDForAssignment(ctx, assignment)
	if err != nil {
		return out, err
	}
	req.RuntimeID = runtimeID
	req.Metadata, err = p.assignmentEventMetadata(ctx, assignment, req.Metadata)
	if err != nil {
		return out, err
	}
	err = p.postJSON(ctx, "/v1/agents/"+url.PathEscape(assignment.AgentID)+"/events", req, &out)
	return out, err
}
