package saasplane

import (
	"context"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneDeliversCancellationFromUnrefreshedHeartbeat(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	first := assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "first",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	}
	fake.enqueue(first)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask first: %v", err)
	}
	if req == nil || req.Metadata[MetadataAssignmentID] != first.ID {
		t.Fatalf("first claim = %+v", req)
	}
	if err := plane.StartTask(context.Background(), req.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	cancelCh, err := plane.WatchCancellation(context.Background(), req.ID)
	if err != nil {
		t.Fatalf("WatchCancellation: %v", err)
	}

	fake.staleHeartbeatIDs[first.ID] = true
	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      RuntimeIDForAgent("daemon-1", AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}),
		RunningTaskIDs: []string{req.ID},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	select {
	case cause := <-cancelCh:
		if cause == nil || !strings.Contains(cause.Error(), "heartbeat lease stale") {
			t.Fatalf("cancel cause = %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for stale heartbeat cancellation")
	}
}
