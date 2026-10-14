package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestPlaneCompleteTaskClassifiesProviderLimitFailure(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-limit",
		TaskID:          "task-limit",
		AgentID:         "agent-research",
		RuntimeProvider: "claude",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-limit",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "agent-research", RuntimeProvider: "claude"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:claude")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if err := plane.CompleteTask(context.Background(), req.ID, agentbridge.Result{
		Status: agentbridge.ResultFailed,
		Error:  "Token quota exceeded while researching",
	}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	last := fake.events[len(fake.events)-1]
	if got := last.Metadata[metadatakeys.AssignmentFailureCategory.String()]; got != "provider_limit" {
		t.Fatalf("failure category metadata = %q, want provider_limit; event=%+v", got, last)
	}
}
