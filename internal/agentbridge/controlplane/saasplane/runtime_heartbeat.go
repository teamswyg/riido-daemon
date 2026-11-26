package saasplane

import (
	"context"
	"errors"
	"net/url"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (p *Plane) Heartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	if p.dynamicBindingsEnabled() {
		return p.dynamicHeartbeat(ctx, hb)
	}
	return p.staticHeartbeat(ctx, hb)
}

func (p *Plane) dynamicHeartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	assignmentsByAgent, err := p.activeAssignmentsForRuntimeHeartbeat(ctx, hb.RuntimeID, hb.RunningTaskIDs)
	if err != nil {
		return err
	}
	var leaseErr error
	for agentID, assignmentIDs := range assignmentsByAgent {
		leaseErr = errors.Join(leaseErr, p.postHeartbeatForAgent(ctx, agentID, hb, assignmentIDs))
	}
	return errors.Join(leaseErr, p.refreshRegisteredRuntimeSnapshot(ctx, hb))
}

func (p *Plane) staticHeartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	agentID, ok := agentFromRuntimeID(hb.RuntimeID)
	if !ok {
		return nil
	}
	assignmentIDs, err := p.activeAssignmentIDsForHeartbeat(ctx, agentID, hb.RunningTaskIDs)
	if err != nil {
		return err
	}
	return p.postHeartbeatForAgent(ctx, agentID, hb, assignmentIDs)
}

func (p *Plane) postHeartbeatForAgent(ctx context.Context, agentID string, hb controlplane.RuntimeHeartbeat, assignmentIDs []string) error {
	if len(assignmentIDs) == 0 {
		return nil
	}
	var out assignmentcontract.AgentHeartbeatResponse
	if err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/heartbeat", heartbeatRequest(p, hb, assignmentIDs), &out); err != nil {
		return err
	}
	return p.deliverUnrefreshedHeartbeatCancels(ctx, assignmentIDs, out)
}

func heartbeatRequest(p *Plane, hb controlplane.RuntimeHeartbeat, assignmentIDs []string) assignmentcontract.AgentHeartbeatRequest {
	return assignmentcontract.AgentHeartbeatRequest{
		DaemonID:            p.cfg.DaemonID,
		DeviceID:            p.cfg.DeviceID,
		RuntimeID:           hb.RuntimeID,
		RunningTaskIDs:      append([]string(nil), hb.RunningTaskIDs...),
		ActiveAssignmentIDs: assignmentIDs,
	}
}
