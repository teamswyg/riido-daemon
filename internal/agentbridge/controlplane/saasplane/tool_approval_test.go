package saasplane

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestResolveToolApprovalCreatesAndWaitsForDecision(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	decidedAt := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	fake.toolDecision = &assignmentcontract.ToolApprovalDecision{
		ApprovalID:   "approval-1",
		AssignmentID: "asn-1",
		Decision:     assignmentcontract.ApprovalDecisionApprove,
		DecidedBy:    "user-1",
		Reason:       "reviewed in web",
		DecidedAt:    decidedAt,
	}
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()
	assignment := assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		State:           assignmentcontract.AssignmentLeased,
	}
	if err := plane.saveAssignmentRuntime(context.Background(), assignment, "daemon-1:codex"); err != nil {
		t.Fatalf("saveAssignmentRuntime: %v", err)
	}

	resolution, err := plane.ResolveToolApproval(context.Background(), "asn-1", agentbridge.ToolRef{
		ID:                "approval-1",
		Name:              "apply_patch",
		Kind:              "patch_apply",
		ProviderRequestID: "req-1",
	})
	if err != nil {
		t.Fatalf("ResolveToolApproval: %v", err)
	}
	if !resolution.Approved || resolution.Reason != "reviewed in web" {
		t.Fatalf("resolution = %+v", resolution)
	}
	if len(fake.toolApprovals) != 1 || fake.toolApprovals[0].RuntimeID != "daemon-1:agent:jykim1:codex" {
		t.Fatalf("tool approvals = %+v", fake.toolApprovals)
	}
	if len(fake.toolApprovalWaits) != 1 || fake.toolApprovalWaits[0].WaitMs <= 0 {
		t.Fatalf("tool approval waits = %+v", fake.toolApprovalWaits)
	}
}
