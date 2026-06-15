package saasplane

import (
	"context"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestTaskRequestKeepsSyntheticDefaultModelIDAsMetadataOnly(t *testing.T) {
	cases := []struct {
		name     string
		provider string
		modelID  string
	}{
		{name: "codex fallback default", provider: "codex", modelID: "codex-default"},
		{name: "claude fallback default", provider: "claude", modelID: "claude-default"},
		{name: "openclaw fallback default", provider: "openclaw", modelID: "openclaw-default"},
		{name: "cursor auto default", provider: "cursor", modelID: "cursor-auto"},
		{name: "unknown fallback default", provider: "other", modelID: "runtime-default"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assignment := assignmentcontract.Assignment{
				ID:              "asn-1",
				TaskID:          "task-a",
				ComponentID:     "component-1",
				AgentID:         "jykim1",
				RuntimeProvider: tc.provider,
				ModelID:         tc.modelID,
				Prompt:          "do the thing",
				LeaseToken:      "lease-1",
			}
			req := taskRequestFromAssignment(assignment)
			if req.Model != "" {
				t.Fatalf("provider model override = %q, want empty for synthetic default", req.Model)
			}
			if got := req.Metadata[MetadataModelID]; got != tc.modelID {
				t.Fatalf("metadata model_id = %q, want %q", got, tc.modelID)
			}
		})
	}
}

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
