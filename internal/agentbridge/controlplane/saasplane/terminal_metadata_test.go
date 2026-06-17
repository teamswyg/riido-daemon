package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestPlaneCompleteTaskReportsBlockedFailureMetadata(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-blocked",
		TaskID:          "task-blocked",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "remove generated files",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-blocked",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if err := plane.CompleteTask(context.Background(), req.ID, agentbridge.Result{
		Status: agentbridge.ResultBlocked,
		Error:  "approval_timeout: no headless approval path",
	}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	last := fake.events[len(fake.events)-1]
	if last.EventType != assignmentcontract.EventAssignmentFailed ||
		last.State != assignmentcontract.AssignmentFailed ||
		last.Message != "approval_timeout: no headless approval path" {
		t.Fatalf("blocked terminal event = %+v", last)
	}
	if got := last.Metadata[metadatakeys.AssignmentResultStatus.String()]; got != string(agentbridge.ResultBlocked) {
		t.Fatalf("result status metadata = %q", got)
	}
	if got := last.Metadata[metadatakeys.AssignmentFailureCategory.String()]; got != "provider_blocked" {
		t.Fatalf("failure category metadata = %q", got)
	}
	if got := last.Metadata[metadatakeys.AssignmentEventKey.String()]; got == "" {
		t.Fatalf("assignment event key metadata missing: %+v", last.Metadata)
	}
}
