package saasplane

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneDetectsCrossAccountConnectionAndRefreshesBindings(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.connectionRevision = "owner-revision"
	fake.connectedPrincipals = 1
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{codexRuntimeBinding("owner-agent")}
	plane := newRuntimeBindingPlane(t, fake, nil)

	initial, err := plane.agentBindings(context.Background())
	if err != nil || len(initial) != 1 {
		t.Fatalf("initial bindings = %+v, %v", initial, err)
	}
	fake.connectionRevision = "cross-account-revision"
	fake.connectedPrincipals = 2
	fake.bindings = append(fake.bindings, codexRuntimeBinding("guest-agent"))
	if err := plane.withState(context.Background(), func(s *planeState) {
		s.agentBindingsCachedAt = time.Now().Add(-agentBindingCacheTTL)
	}); err != nil {
		t.Fatalf("expire binding cache: %v", err)
	}

	refreshed, err := plane.agentBindings(context.Background())
	if err != nil || len(refreshed) != 2 {
		t.Fatalf("refreshed bindings = %+v, %v", refreshed, err)
	}
	if !daemonBindingContainsAgent(refreshed, "guest-agent") {
		t.Fatalf("cross-account agent not detected: %+v", refreshed)
	}
	var revision string
	var principalCount int
	if err := plane.withState(context.Background(), func(s *planeState) {
		revision = s.connectionRevision
		principalCount = s.connectedPrincipalCount
	}); err != nil {
		t.Fatalf("read observation state: %v", err)
	}
	if revision != "cross-account-revision" || principalCount != 2 {
		t.Fatalf("connection observation revision=%q principals=%d", revision, principalCount)
	}
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-cross-account",
		TaskID:          "task-cross-account",
		AgentID:         "guest-agent",
		RuntimeProvider: "codex",
		State:           assignmentcontract.AssignmentQueued,
	})
	claimed, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil || claimed == nil || claimed.ID != "asn-cross-account" {
		t.Fatalf("cross-account claim = %+v, %v", claimed, err)
	}
}

func daemonBindingContainsAgent(bindings []assignmentcontract.AgentRuntimeBinding, agentID string) bool {
	for _, binding := range bindings {
		if binding.AgentID == agentID {
			return true
		}
	}
	return false
}
