package saasplane

import (
	"context"
	"fmt"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (p *Plane) activeAssignmentsByAgentForHeartbeat(ctx context.Context, runningTaskIDs []string) (map[string][]string, error) {
	executions := normalizedExecutionIDs(runningTaskIDs)
	if len(executions) == 0 {
		return nil, nil
	}
	byAgent := map[string][]string{}
	err := p.withState(ctx, func(s *planeState) {
		for _, executionID := range executions {
			assignment := s.assignmentsByExecution[executionID]
			if assignment.AgentID == "" || assignment.ID == "" {
				continue
			}
			byAgent[assignment.AgentID] = append(byAgent[assignment.AgentID], assignment.ID)
		}
	})
	return byAgent, err
}

func (p *Plane) deliverCancel(ctx context.Context, assignment assignmentcontract.Assignment) error {
	return p.withState(ctx, func(s *planeState) {
		sendAndCloseCancelWatcher(s, assignmentExecutionID(assignment), fmt.Errorf("saas assignment %s cancelled", assignment.ID))
	})
}

func (p *Plane) deliverUnrefreshedHeartbeatCancels(ctx context.Context, requestedAssignmentIDs []string, response assignmentcontract.AgentHeartbeatResponse) error {
	if len(requestedAssignmentIDs) == 0 {
		return nil
	}
	refreshed := refreshedAssignmentIDs(response.RefreshedAssignments)
	return p.withState(ctx, func(s *planeState) {
		cancelUnrefreshedHeartbeatAssignments(s, requestedAssignmentIDs, refreshed)
	})
}

func refreshedAssignmentIDs(assignments []assignmentcontract.Assignment) map[string]bool {
	refreshed := map[string]bool{}
	for _, assignment := range assignments {
		if strings.TrimSpace(assignment.ID) != "" {
			refreshed[assignment.ID] = true
		}
	}
	return refreshed
}

func cancelUnrefreshedHeartbeatAssignments(s *planeState, requestedAssignmentIDs []string, refreshed map[string]bool) {
	for _, assignmentID := range requestedAssignmentIDs {
		assignmentID = strings.TrimSpace(assignmentID)
		if assignmentID == "" || refreshed[assignmentID] {
			continue
		}
		if assignment, ok := s.assignmentsByExecution[assignmentID]; ok {
			sendAndCloseCancelWatcher(s, assignmentID, fmt.Errorf("saas assignment %s heartbeat lease stale", assignment.ID))
			delete(s.assignmentsByExecution, assignmentID)
			delete(s.runtimeIDsByExecution, assignmentID)
			delete(s.partialBodies, assignmentID)
		}
	}
}
