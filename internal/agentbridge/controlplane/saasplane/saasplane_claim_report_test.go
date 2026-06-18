package saasplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneClaimsAndReportsAssignment(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	assignment := saasClaimReportAssignment()
	fake.enqueue(assignment)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	assertClaimReportRequest(t, req, assignment)

	if err := reportClaimedAssignmentLifecycle(context.Background(), plane, req); err != nil {
		t.Fatal(err)
	}
	assertClaimReportEvents(t, fake)

	heartbeats := fake.heartbeatsFor("jykim1")
	if len(heartbeats) != 1 ||
		len(heartbeats[0].ActiveAssignmentIDs) != 1 ||
		heartbeats[0].ActiveAssignmentIDs[0] != assignment.ID {
		t.Fatalf("heartbeats = %+v", heartbeats)
	}
	if got := req.Metadata[controlplane.MetadataTaskID]; got != assignment.TaskID {
		t.Fatalf("task metadata = %q", got)
	}
}
