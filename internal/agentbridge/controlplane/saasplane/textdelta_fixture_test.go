package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func newTextDeltaFixture(t *testing.T) (*fakeAssignmentServer, *Plane, *bridge.TaskRequest) {
	t.Helper()
	fake := newFakeAssignmentServer(t)
	fake.enqueue(textDeltaAssignment())

	plane := newTestPlane(t, fake.URL(), []AgentBinding{textDeltaAgentBinding()})
	t.Cleanup(plane.Close)

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	return fake, plane, req
}

func textDeltaAssignment() assignmentcontract.Assignment {
	return assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	}
}

func textDeltaAgentBinding() AgentBinding {
	return AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}
}

func reportTextDeltas(t *testing.T, plane *Plane, executionID string, deltas ...string) {
	t.Helper()
	for _, delta := range deltas {
		if err := plane.ReportEvent(context.Background(), executionID, agentbridge.Event{
			Kind: agentbridge.EventTextDelta,
			Text: delta,
		}); err != nil {
			t.Fatalf("ReportEvent text delta: %v", err)
		}
	}
}
