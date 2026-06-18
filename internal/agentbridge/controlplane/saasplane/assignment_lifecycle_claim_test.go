package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func claimSameTaskAssignments(
	t *testing.T,
	plane *Plane,
	first assignmentcontract.Assignment,
	second assignmentcontract.Assignment,
) (*bridge.TaskRequest, *bridge.TaskRequest) {
	t.Helper()
	req1 := claimLifecycleAssignment(t, plane, "first")
	req2 := claimLifecycleAssignment(t, plane, "second")
	if req1.ID != first.ID || req2.ID != second.ID {
		t.Fatalf("claims = %+v / %+v", req1, req2)
	}
	if req1.Metadata[controlplane.MetadataTaskID] != first.TaskID {
		t.Fatalf("first logical task metadata lost: %+v", req1.Metadata)
	}
	if req2.Metadata[controlplane.MetadataTaskID] != second.TaskID {
		t.Fatalf("second logical task metadata lost: %+v", req2.Metadata)
	}
	return req1, req2
}

func claimLifecycleAssignment(t *testing.T, plane *Plane, label string) *bridge.TaskRequest {
	t.Helper()
	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask %s: %v", label, err)
	}
	if req == nil {
		t.Fatalf("ClaimTask %s returned nil", label)
	}
	return req
}

func startSameTaskAssignments(t *testing.T, plane *Plane, firstID, secondID string) {
	t.Helper()
	if err := plane.StartTask(context.Background(), firstID); err != nil {
		t.Fatalf("StartTask first: %v", err)
	}
	if err := plane.StartTask(context.Background(), secondID); err != nil {
		t.Fatalf("StartTask second: %v", err)
	}
}
