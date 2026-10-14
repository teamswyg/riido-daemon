package saasplane

import (
	"encoding/json"
	"net/http"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (f *fakeAssignmentServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req assignmentcontract.AgentHeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f.heartbeats = append(f.heartbeats, req)
	writeJSON(w, assignmentcontract.AgentHeartbeatResponse{
		SchemaVersion:        assignmentcontract.SchemaVersion,
		RefreshedAssignments: f.refreshedAssignments(req.ActiveAssignmentIDs),
	})
}

func (f *fakeAssignmentServer) refreshedAssignments(ids []string) []assignmentcontract.Assignment {
	var refreshed []assignmentcontract.Assignment
	for _, assignmentID := range ids {
		if f.staleHeartbeatIDs[assignmentID] {
			continue
		}
		assignment, ok := f.assignmentsByID[assignmentID]
		if ok {
			refreshed = append(refreshed, assignment)
		}
	}
	return refreshed
}
