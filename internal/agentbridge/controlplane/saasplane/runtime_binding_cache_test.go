package saasplane

import (
	"context"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneCachesDynamicAgentBindingsAcrossClaimWave(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{
		codexRuntimeBinding("agent-codex"),
		cursorRuntimeBinding("agent-cursor"),
	}
	plane := newRuntimeBindingPlane(t, fake, nil)

	for _, runtimeID := range []string{"daemon-1:codex", "daemon-1:cursor"} {
		req, err := plane.ClaimTask(context.Background(), runtimeID)
		if err != nil {
			t.Fatalf("ClaimTask %s: %v", runtimeID, err)
		}
		if req != nil {
			t.Fatalf("empty queue should not claim task: %+v", req)
		}
	}
	assertBindingCacheClaimWave(t, fake)
	registerRuntimeForBinding(t, plane, controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:codex",
		Provider:  "codex",
	})
	if _, err := plane.ClaimTask(context.Background(), "daemon-1:codex"); err != nil {
		t.Fatalf("ClaimTask after runtime snapshot: %v", err)
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 2 {
		t.Fatalf("agent-bindings request count after runtime snapshot = %d, want 2", got)
	}
}

func TestPlaneCachesEmptyDynamicAgentBindingsAcrossClaimWave(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newRuntimeBindingPlane(t, fake, nil)
	for _, runtimeID := range []string{"daemon-1:codex", "daemon-1:cursor"} {
		req, err := plane.ClaimTask(context.Background(), runtimeID)
		if err != nil {
			t.Fatalf("ClaimTask %s: %v", runtimeID, err)
		}
		if req != nil {
			t.Fatalf("empty binding list should not claim task: %+v", req)
		}
	}
	if got := fake.requestCount("/v1/daemon/agent-bindings"); got != 1 {
		t.Fatalf("empty agent-bindings request count = %d, want 1", got)
	}
}
