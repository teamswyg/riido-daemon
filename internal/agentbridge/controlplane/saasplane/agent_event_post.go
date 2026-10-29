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
	binding, err := p.claimBinding(ctx, assignmentExecutionID(assignment))
	if err != nil {
		return out, err
	}
	if binding != nil {
		req.DeviceID, req.DaemonID = binding.deviceID, binding.daemonID
		runtimeID = binding.runtimeID
	}
	req.RuntimeID = runtimeID
	req.Metadata, err = p.assignmentEventMetadata(ctx, assignment, req.Metadata)
	if err != nil {
		return out, err
	}
	err = p.postJSON(ctx, "/v1/agents/"+url.PathEscape(assignment.AgentID)+"/events", req, &out)
	if err == nil {
		err = p.withState(ctx, func(s *planeState) {
			current, ok := s.assignmentsByExecution[assignmentExecutionID(assignment)]
			if !ok || current.ID != assignment.ID || current.LeaseToken != assignment.LeaseToken || !currentClaimScope(ctx, s, assignmentExecutionID(assignment)) {
				return
			}
			if req.EventType == assignmentcontract.EventAssignmentWaitingForUser {
				current.ConfirmationReason = req.ConfirmationReason
			} else if req.EventType == assignmentcontract.EventAssignmentRunning || assignmentcontract.IsTerminal(req.State) {
				current.ConfirmationReason = ""
			}
			s.assignmentsByExecution[assignmentExecutionID(assignment)] = current
		})
	}
	return out, err
}
