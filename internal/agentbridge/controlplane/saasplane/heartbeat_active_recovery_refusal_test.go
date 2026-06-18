package saasplane

import (
	"context"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestPlaneFailsActiveAssignmentWithoutSessionAfterLocalStateLoss(t *testing.T) {
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
	if len(fake.events) != 1 {
		t.Fatalf("events = %+v, want one recovery failure event", fake.events)
	}
	event := fake.events[0]
	if event.EventType != assignmentcontract.EventAssignmentFailed ||
		event.State != assignmentcontract.AssignmentFailed ||
		event.Metadata[metadatakeys.AssignmentRecovery.String()] != assignmentcontract.RecoveryFreshStartRefused.String() {
		t.Fatalf("recovery failure event = %+v", event)
	}
	if !strings.Contains(event.Message, "refusing fresh start") {
		t.Fatalf("recovery failure message = %q", event.Message)
	}
}
