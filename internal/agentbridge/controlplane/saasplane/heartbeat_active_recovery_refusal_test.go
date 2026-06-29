package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneSkipsActiveAssignmentWithoutSessionAfterLocalStateLoss(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	active := assignmentcontract.Assignment{
		ID:              "asn-active",
		TaskID:          "task-active",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "would duplicate side effects if started fresh",
		State:           assignmentcontract.AssignmentRunning,
		LeaseToken:      "lease-active",
	}
	fake.activeNext(active.AgentID, active)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask active without session: %v", err)
	}
	if req != nil {
		t.Fatalf("active assignment without session must not be fresh-started: %+v", req)
	}
	if len(fake.events) != 0 {
		t.Fatalf("events = %+v, want no terminal event from a stateless poller", fake.events)
	}
}
