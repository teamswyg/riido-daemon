package saasplane

import assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

func (f *fakeAssignmentServer) enqueue(assignment assignmentcontract.Assignment) {
	f.assignmentsByAgent[assignment.AgentID] = append(f.assignmentsByAgent[assignment.AgentID], assignment)
	f.assignmentsByID[assignment.ID] = assignment
}

func (f *fakeAssignmentServer) cancelNext(agentID string, assignment assignmentcontract.Assignment) {
	assignment.State = assignmentcontract.AssignmentCancelling
	f.cancelByAgent[agentID] = assignment
	f.assignmentsByID[assignment.ID] = assignment
}

func (f *fakeAssignmentServer) activeNext(agentID string, assignment assignmentcontract.Assignment) {
	if assignment.State == "" {
		assignment.State = assignmentcontract.AssignmentLeased
	}
	f.activeByAgent[agentID] = assignment
	f.assignmentsByID[assignment.ID] = assignment
}

func (f *fakeAssignmentServer) failNext(path string, count, status int) {
	f.transientFailures[path] = count
	f.transientStatuses[path] = status
}

func (f *fakeAssignmentServer) requestCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requestCounts[path]
}

func (f *fakeAssignmentServer) pollRequestsFor(agentID string) []assignmentcontract.PollRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]assignmentcontract.PollRequest(nil), f.pollRequestsByAgent[agentID]...)
}
