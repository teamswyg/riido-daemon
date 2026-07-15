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
	fake.replaceBindingsAfterNextFailure(
		"/v1/agents/agent-stale-codex/poll",
		[]assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("agent-live-codex")},
	)
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
	if got := len(fake.pollRequestsFor("agent-stale-codex")); got != 0 {
		t.Fatalf("stale candidate successful poll count = %d, want 0", got)
	}
	if got := fake.requestCount("/v1/agents/agent-stale-codex/poll"); got != 1 {
		t.Fatalf("stale candidate request count = %d, want 1", got)
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 2 {
		t.Fatalf("agent-bindings request count = %d, want initial plus reconciliation", got)
	}
}

func TestPlaneSuppressesRejectedDynamicCandidateUntilBindingCacheRefresh(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("agent-stale-codex")}
	fake.failNext("/v1/agents/agent-stale-codex/poll", 2, http.StatusBadRequest)
	plane := newRuntimeBindingPlane(t, fake, nil)

	for i := range 2 {
		req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
		if err != nil {
			t.Fatalf("ClaimTask %d: %v", i+1, err)
		}
		if req != nil {
			t.Fatalf("ClaimTask %d request = %+v, want nil", i+1, req)
		}
	}
	if got := fake.requestCount("/v1/agents/agent-stale-codex/poll"); got != 1 {
		t.Fatalf("stale candidate request count = %d, want one bounded rejection", got)
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 2 {
		t.Fatalf("agent-bindings request count = %d, want initial plus reconciliation", got)
	}
}
