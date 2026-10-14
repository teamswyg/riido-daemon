package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlanePollsOnlyRuntimeScopedAgent(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-2",
		TaskID:          "task-b",
		ComponentID:     "component-1",
		AgentID:         "jykim2",
		RuntimeProvider: "codex",
		Prompt:          "second agent task",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-2",
	})
	agents := []AgentBinding{
		{AgentID: "jykim1", RuntimeProvider: "codex"},
		{AgentID: "jykim2", RuntimeProvider: "codex"},
	}
	plane := newTestPlane(t, fake.URL(), agents)
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), RuntimeIDForAgent("daemon-1", agents[0]))
	if err != nil {
		t.Fatalf("ClaimTask jykim1: %v", err)
	}
	if req != nil {
		t.Fatalf("jykim1 runtime claimed another agent task: %+v", req)
	}
	req, err = plane.ClaimTask(context.Background(), RuntimeIDForAgent("daemon-1", agents[1]))
	if err != nil {
		t.Fatalf("ClaimTask jykim2: %v", err)
	}
	if req == nil || req.ID != "asn-2" || req.Metadata[MetadataAgentID] != "jykim2" {
		t.Fatalf("jykim2 claim = %+v", req)
	}
}
