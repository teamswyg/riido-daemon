package saasplane

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (f *fakeAssignmentServer) handleEvents(w http.ResponseWriter, r *http.Request, agentID string) {
	var req assignmentcontract.AgentEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.AssignmentID == "" || req.TaskID == "" || req.DaemonID == "" || req.RuntimeID == "" {
		http.Error(w, "missing event identity", http.StatusBadRequest)
		return
	}
	f.events = append(f.events, req)
	f.applyEventState(req)
	writeJSON(w, assignmentcontract.AgentEventResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
		Assignment:    func() *assignmentcontract.Assignment { value := f.assignmentsByID[req.AssignmentID]; return &value }(),
		Event: assignmentcontract.TaskEvent{
			WorkspaceID:  f.assignmentsByID[req.AssignmentID].WorkspaceID,
			TeamID:       f.assignmentsByID[req.AssignmentID].TeamID,
			CreatedBy:    f.assignmentsByID[req.AssignmentID].CreatedBy,
			EventID:      fmt.Sprintf("%s:%d", req.AssignmentID, len(f.events)),
			Seq:          int64(len(f.events)),
			TaskID:       req.TaskID,
			AssignmentID: req.AssignmentID,
			AgentID:      agentID,
			Type:         req.EventType,
			State:        req.State,
			Message:      req.Message,
			Metadata: func() map[string]string {
				metadata := cloneEventMetadata(req.Metadata)
				delete(metadata, "lease_token")
				return metadata
			}(),
			At: time.Now().UTC(),
		},
	})
}

func (f *fakeAssignmentServer) applyEventState(req assignmentcontract.AgentEventRequest) {
	assignment, ok := f.assignmentsByID[req.AssignmentID]
	if ok && req.State != "" {
		assignment.State = req.State
		f.assignmentsByID[req.AssignmentID] = assignment
	}
}
