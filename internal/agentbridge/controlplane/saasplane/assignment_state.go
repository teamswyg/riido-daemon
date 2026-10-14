package saasplane

import (
	"context"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (p *Plane) saveAssignmentRuntime(ctx context.Context, assignment assignmentcontract.Assignment, runtimeID string) error {
	return p.withState(ctx, func(s *planeState) {
		executionID := assignmentExecutionID(assignment)
		s.assignmentsByExecution[executionID] = assignment
		if runtimeID != "" {
			s.runtimeIDsByExecution[executionID] = runtimeID
		}
	})
}

func (p *Plane) assignmentForExecution(ctx context.Context, executionID string) (assignmentcontract.Assignment, bool, error) {
	var assignment assignmentcontract.Assignment
	var ok bool
	err := p.withState(ctx, func(s *planeState) {
		assignment, ok = s.assignmentsByExecution[executionID]
	})
	return assignment, ok, err
}

func (p *Plane) activeAssignmentIDsForHeartbeat(ctx context.Context, agentID string, runningTaskIDs []string) ([]string, error) {
	executions := uniqueExecutionIDs(runningTaskIDs)
	if len(executions) == 0 {
		return nil, nil
	}
	var ids []string
	err := p.withState(ctx, func(s *planeState) {
		for _, executionID := range executions {
			assignment := s.assignmentsByExecution[executionID]
			if assignment.AgentID != agentID || assignment.ID == "" {
				continue
			}
			ids = append(ids, assignment.ID)
		}
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func uniqueExecutionIDs(runningTaskIDs []string) []string {
	var executions []string
	seen := map[string]bool{}
	for _, executionID := range runningTaskIDs {
		executionID = strings.TrimSpace(executionID)
		if executionID == "" || seen[executionID] {
			continue
		}
		seen[executionID] = true
		executions = append(executions, executionID)
	}
	return executions
}
