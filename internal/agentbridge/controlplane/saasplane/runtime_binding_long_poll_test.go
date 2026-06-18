package saasplane

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneLongPollsOnlyOneDynamicCandidatePerRuntime(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bindings = []assignmentcontract.AgentRuntimeBinding{
		codexRuntimeBinding("agent-riido-codex"),
		codexRuntimeBinding("agent-youngsil-codex"),
	}
	plane := newRuntimeBindingPlane(t, fake, func(cfg *Config) {
		cfg.LongPollWait = 2500 * time.Millisecond
	})

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req != nil {
		t.Fatalf("empty queue should not claim task: %+v", req)
	}
	riidoPolls := fake.pollRequestsFor("agent-riido-codex")
	youngsilPolls := fake.pollRequestsFor("agent-youngsil-codex")
	if len(riidoPolls) != 2 || len(youngsilPolls) != 1 {
		t.Fatalf("poll requests riido=%+v youngsil=%+v", riidoPolls, youngsilPolls)
	}
	if riidoPolls[0].WaitMs != 0 || riidoPolls[1].WaitMs != 2500 || youngsilPolls[0].WaitMs != 0 {
		t.Fatalf("unexpected wait_ms riido=%+v youngsil=%+v", riidoPolls, youngsilPolls)
	}
}
