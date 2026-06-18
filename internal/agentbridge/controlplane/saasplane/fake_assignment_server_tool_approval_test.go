package saasplane

import (
	"encoding/json"
	"net/http"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (f *fakeAssignmentServer) handleToolApprovals(w http.ResponseWriter, r *http.Request, _ string, parts []string) {
	switch {
	case len(parts) == 0:
		f.handleToolApprovalCreate(w, r)
	case len(parts) == 2 && parts[1] == "wait":
		f.handleToolApprovalWait(w, r, parts[0])
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeAssignmentServer) handleToolApprovalCreate(w http.ResponseWriter, r *http.Request) {
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
}

func (f *fakeAssignmentServer) handleToolApprovalWait(w http.ResponseWriter, r *http.Request, approvalID string) {
	var req assignmentcontract.ToolApprovalWaitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f.toolApprovalWaits = append(f.toolApprovalWaits, req)
	writeJSON(w, assignmentcontract.ToolApprovalWaitResponse{
		SchemaVersion: assignmentcontract.SchemaVersion,
		Result:        f.toolApprovalResult(req, approvalID),
		Decision:      f.toolDecision,
	})
}

func (f *fakeAssignmentServer) toolApprovalResult(
	req assignmentcontract.ToolApprovalWaitRequest,
	approvalID string,
) assignmentcontract.ToolApprovalResult {
	status := f.toolApprovalStatus
	if status == "" {
		status = assignmentcontract.ApprovalApproved
	}
	return assignmentcontract.ToolApprovalResult{
		ApprovalID:   approvalID,
		AssignmentID: req.AssignmentID,
		Status:       status,
	}
}
