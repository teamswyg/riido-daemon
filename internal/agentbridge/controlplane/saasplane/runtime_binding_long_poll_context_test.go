package saasplane

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneDisablesDynamicLongPollFromClaimContext(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{
		codexRuntimeBinding("agent-riido-codex"),
		codexRuntimeBinding("agent-youngsil-codex"),
	}
	plane := newRuntimeBindingPlane(t, fake, func(cfg *Config) {
		cfg.LongPollWait = 2500 * time.Millisecond
	})
	ctx := controlplane.ContextWithClaimLongPoll(context.Background(), false)

	req, err := plane.ClaimTask(ctx, "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req != nil {
		t.Fatalf("empty queue should not claim task: %+v", req)
	}
	assertNoDynamicLongPoll(t, fake)
}

func assertNoDynamicLongPoll(t *testing.T, fake *fakeAssignmentServer) {
	t.Helper()
	for _, agentID := range []string{"agent-riido-codex", "agent-youngsil-codex"} {
		polls := fake.pollRequestsFor(agentID)
		if len(polls) != 1 || polls[0].WaitMs != 0 {
			t.Fatalf("poll requests for %s = %+v, want one short poll", agentID, polls)
		}
	}
}
