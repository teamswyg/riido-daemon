package saasplane

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

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
	if cancel, ok := f.cancelByAgent[agentID]; ok {
		delete(f.cancelByAgent, agentID)
		f.assignmentsByID[cancel.ID] = cancel
		writeJSON(w, assignmentcontract.PollResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Action:        assignmentcontract.PollCancel,
			Assignment:    &cancel,
		})
		return
	}
	if active, ok := f.activeByAgent[agentID]; ok {
		delete(f.activeByAgent, agentID)
		f.assignmentsByID[active.ID] = active
		writeJSON(w, assignmentcontract.PollResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Action:        assignmentcontract.PollActive,
			Assignment:    &active,
		})
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
	f.assignmentsByID[assignment.ID] = assignment
	writeJSON(w, assignmentcontract.PollResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
		Action:        assignmentcontract.PollStart,
		Assignment:    &assignment,
	})
}

func (f *fakeAssignmentServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req assignmentcontract.AgentHeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f.heartbeats = append(f.heartbeats, req)
	var refreshed []assignmentcontract.Assignment
	for _, assignmentID := range req.ActiveAssignmentIDs {
		if f.staleHeartbeatIDs[assignmentID] {
			continue
		}
		assignment, ok := f.assignmentsByID[assignmentID]
		if !ok {
			continue
		}
		refreshed = append(refreshed, assignment)
	}
	writeJSON(w, assignmentcontract.AgentHeartbeatResponse{
		SchemaVersion:        assignmentcontract.SchemaVersion,
		RefreshedAssignments: refreshed,
	})
}

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
	if assignment, ok := f.assignmentsByID[req.AssignmentID]; ok && req.State != "" {
		assignment.State = req.State
		f.assignmentsByID[req.AssignmentID] = assignment
	}
	writeJSON(w, assignmentcontract.AgentEventResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
		Event: assignmentcontract.TaskEvent{
			Seq:          int64(len(f.events)),
			TaskID:       req.TaskID,
			AssignmentID: req.AssignmentID,
			AgentID:      agentID,
			Type:         req.EventType,
			State:        req.State,
			Message:      req.Message,
			At:           time.Now().UTC(),
		},
	})
}

func (f *fakeAssignmentServer) handleToolApprovals(w http.ResponseWriter, r *http.Request, _ string, parts []string) {
	switch {
	case len(parts) == 0:
		var req assignmentcontract.ToolApprovalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.toolApprovals = append(f.toolApprovals, req)
		writeJSON(w, assignmentcontract.ToolApprovalCreateResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Approval:      req,
		})
	case len(parts) == 2 && parts[1] == "wait":
		var req assignmentcontract.ToolApprovalWaitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.toolApprovalWaits = append(f.toolApprovalWaits, req)
		status := f.toolApprovalStatus
		if status == "" {
			status = assignmentcontract.ApprovalApproved
		}
		writeJSON(w, assignmentcontract.ToolApprovalWaitResponse{
			SchemaVersion: assignmentcontract.SchemaVersion,
			Result: assignmentcontract.ToolApprovalResult{
				ApprovalID:   parts[0],
				AssignmentID: req.AssignmentID,
				Status:       status,
			},
			Decision: f.toolDecision,
		})
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeAssignmentServer) assertEvent(t *testing.T, eventType string) {
	t.Helper()
	for _, ev := range f.events {
		if ev.EventType == eventType {
			return
		}
	}
	t.Fatalf("event %q missing from %+v", eventType, f.events)
}

func (f *fakeAssignmentServer) heartbeatsFor(agentID string) []assignmentcontract.AgentHeartbeatRequest {
	var out []assignmentcontract.AgentHeartbeatRequest
	for _, hb := range f.heartbeats {
		if runtimeAgent, ok := agentFromRuntimeID(hb.RuntimeID); ok && runtimeAgent == agentID {
			out = append(out, hb)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, value any) {
	_ = json.NewEncoder(w).Encode(value)
}
