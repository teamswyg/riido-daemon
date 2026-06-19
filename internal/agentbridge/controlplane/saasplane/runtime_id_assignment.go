package saasplane

import (
	"context"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func (p *Plane) runtimeIDForAssignment(ctx context.Context, assignment assignmentcontract.Assignment) (string, error) {
	if p.dynamicBindingsEnabled() {
		runtimeID, err := p.savedRuntimeIDForAssignment(ctx, assignment)
		if err != nil || strings.TrimSpace(runtimeID) != "" {
			return runtimeID, err
		}
	}
	return p.configuredRuntimeIDForAssignment(assignment), nil
}

func (p *Plane) savedRuntimeIDForAssignment(ctx context.Context, assignment assignmentcontract.Assignment) (string, error) {
	var runtimeID string
	err := p.withState(ctx, func(s *planeState) {
		runtimeID = s.runtimeIDsByExecution[assignmentExecutionID(assignment)]
	})
	return runtimeID, err
}

func (p *Plane) configuredRuntimeIDForAssignment(assignment assignmentcontract.Assignment) string {
	for _, agent := range p.cfg.Agents {
		if agentMatchesAssignmentRuntime(agent, assignment) {
			return RuntimeIDForAgent(p.cfg.DaemonID, agent)
		}
	}
	return RuntimeIDForAgent(p.cfg.DaemonID, AgentBinding{
		AgentID:         assignment.AgentID,
		RuntimeProvider: assignment.RuntimeProvider,
	})
}

func agentMatchesAssignmentRuntime(agent AgentBinding, assignment assignmentcontract.Assignment) bool {
	return agent.AgentID == assignment.AgentID &&
		providercatalog.Normalize(agent.RuntimeProvider) == providercatalog.Normalize(assignment.RuntimeProvider)
}

func assignmentExecutionID(assignment assignmentcontract.Assignment) string {
	return assignmentcontract.ExecutionIDFromAssignment(assignment)
}
