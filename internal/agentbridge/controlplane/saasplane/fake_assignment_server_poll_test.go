package saasplane

import (
	"encoding/json"
	"net/http"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (f *fakeAssignmentServer) handlePoll(w http.ResponseWriter, r *http.Request, agentID string) {
	var req assignmentcontract.PollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.DaemonID == "" || req.RuntimeID == "" {
		http.Error(w, "missing poll identity", http.StatusBadRequest)
		return
	}
	f.pollRequestsByAgent[agentID] = append(f.pollRequestsByAgent[agentID], req)
	f.writePollResponse(w, agentID)
}

func (f *fakeAssignmentServer) writePollResponse(w http.ResponseWriter, agentID string) {
	if cancel, ok := f.cancelByAgent[agentID]; ok {
		delete(f.cancelByAgent, agentID)
		f.writeAssignmentPoll(w, assignmentcontract.PollCancel, cancel)
		return
	}
	if active, ok := f.activeByAgent[agentID]; ok {
		delete(f.activeByAgent, agentID)
		f.writeAssignmentPoll(w, assignmentcontract.PollActive, active)
		return
	}
	queue := f.assignmentsByAgent[agentID]
	if len(queue) == 0 {
		writeJSON(w, assignmentcontract.PollResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Action:        assignmentcontract.PollNone,
		})
		return
	}
	assignment := queue[0]
	f.assignmentsByAgent[agentID] = queue[1:]
	assignment.State = assignmentcontract.AssignmentLeased
	f.writeAssignmentPoll(w, assignmentcontract.PollStart, assignment)
}

func (f *fakeAssignmentServer) writeAssignmentPoll(
	w http.ResponseWriter,
	action assignmentcontract.PollAction,
	assignment assignmentcontract.Assignment,
) {
	f.assignmentsByID[assignment.ID] = assignment
	writeJSON(w, assignmentcontract.PollResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
		Action:        action,
		Assignment:    &assignment,
	})
}
