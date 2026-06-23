package saasplane

import (
	"context"
	"net/http"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneSkipsStaleDynamicCandidateAndClaimsLiveAgent(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{
		codexRuntimeBinding("agent-stale-codex"),
		codexRuntimeBinding("agent-live-codex"),
	}
	fake.failNext("/v1/agents/agent-stale-codex/poll", 1, http.StatusBadRequest)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-live",
		TaskID:          "task-live",
		ComponentID:     "component-live",
		AgentID:         "agent-live-codex",
		RuntimeProvider: "codex",
		Prompt:          "use provider document context",
		State:           assignmentcontract.AssignmentQueued,
	})
	plane := newRuntimeBindingPlane(t, fake, nil)

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req == nil || req.ID != "asn-live" {
		t.Fatalf("ClaimTask request = %+v, want asn-live", req)
	}
	if got := len(fake.pollRequestsFor("agent-live-codex")); got == 0 {
		t.Fatal("live candidate was not polled after stale binding rejection")
	}
}
