package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneSendsBearerToken(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bearerToken = "secret"
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane := newTestPlaneWithToken(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}}, "secret")
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask with token: %v", err)
	}
	if req == nil || req.ID != "asn-1" {
		t.Fatalf("request = %+v", req)
	}
}
