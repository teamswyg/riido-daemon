package saasplane

import (
	"context"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneDeliversCancellationFromPollResponse(t *testing.T) {
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

	fake.cancelNext(first.AgentID, first)
	req, err = plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask cancel poll: %v", err)
	}
	if req != nil {
		t.Fatalf("cancel poll should not claim new task: %+v", req)
	}
	select {
	case cause := <-cancelCh:
		if cause == nil || !strings.Contains(cause.Error(), first.ID) {
			t.Fatalf("cancel cause = %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for cancellation")
	}
}
