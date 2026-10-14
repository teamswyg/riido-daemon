package taskdbplane

import (
	"context"
	"testing"
)

func TestClaimTaskReusesExistingApprovalIDForHumanGatedTask(t *testing.T) {
	path := writeTaskDB(t, approvedHumanClaimDB(t))
	plane := newTestPlane(t, path)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.ID != "task-human" {
		t.Fatalf("unexpected request: %+v", req)
	}
	loaded := loadTaskDB(t, path)
	last := loaded.CommandReceipts[len(loaded.CommandReceipts)-1]
	if last.ApprovalID != "approval:human:1" || !last.RequiresHumanApproval {
		t.Fatalf("claim receipt did not reuse approval: %+v", last)
	}
}
