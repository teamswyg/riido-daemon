package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneClaimsActiveAssignmentAfterLocalStateLoss(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	active := assignmentcontract.Assignment{
		ID:                       "asn-active",
		TaskID:                   "task-active",
		ComponentID:              "component-1",
		AgentID:                  "jykim1",
		RuntimeProvider:          "codex",
		Prompt:                   "resume active assignment",
		State:                    assignmentcontract.AssignmentLeased,
		LeaseToken:               "lease-active",
		AllowExperimentalRuntime: true,
		ResumeSessionID:          "sess-initial",
		ProviderSessionID:        "sess-current",
	}
	fake.activeNext(active.AgentID, active)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask active: %v", err)
	}
	if req == nil || req.ID != active.ID || req.Metadata[MetadataAssignmentID] != active.ID {
		t.Fatalf("active claim = %+v", req)
	}
	if !req.AllowExperimentalRuntime {
		t.Fatal("active assignment should preserve experimental opt-in")
	}
	if req.ResumeSessionID != active.ProviderSessionID {
		t.Fatalf("active assignment resume_session_id = %q, want provider session %q", req.ResumeSessionID, active.ProviderSessionID)
	}
}
