package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneClaimsDynamicAgentBinding(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("jykim1")}
	fake.enqueue(queuedHTTPAssignment("asn-1", "dynamic binding task"))
	plane := newRuntimeBindingPlane(t, fake, nil)

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req == nil || req.ID != "asn-1" || req.Provider != "codex" || req.Metadata[MetadataAgentID] != "jykim1" {
		t.Fatalf("dynamic claim = %+v", req)
	}
	if err := plane.StartTask(context.Background(), req.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	err = plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      "daemon-1:codex",
		RunningTaskIDs: []string{req.ID},
	})
	if err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	assertDynamicHeartbeat(t, fake, req.ID)
	err = plane.ReportEvent(context.Background(), req.ID, agentbridge.Event{
		Kind: agentbridge.EventProgress,
		Text: "dynamic progress",
	})
	if err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if len(fake.events) < 2 || fake.events[len(fake.events)-1].RuntimeID != "daemon-1:codex" {
		t.Fatalf("dynamic events = %+v", fake.events)
	}
}
