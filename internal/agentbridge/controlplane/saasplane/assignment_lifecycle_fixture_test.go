package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func setupSameTaskAssignments(t *testing.T) (
	*fakeAssignmentServer,
	*Plane,
	assignmentcontract.Assignment,
	assignmentcontract.Assignment,
) {
	t.Helper()
	fake := newFakeAssignmentServer(t)
	first := assignmentLifecycleFixture("asn-1", "first", "lease-1")
	second := assignmentLifecycleFixture("asn-2", "second", "lease-2")
	fake.enqueue(first)
	fake.enqueue(second)
	return fake, newLifecyclePlane(t, fake), first, second
}

func assignmentLifecycleFixture(id, prompt, leaseToken string) assignmentcontract.Assignment {
	return assignmentcontract.Assignment{
		ID:              id,
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          prompt,
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      leaseToken,
	}
}

func newLifecyclePlane(t *testing.T, fake *fakeAssignmentServer) *Plane {
	t.Helper()
	agents := []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}}
	return newTestPlane(t, fake.URL(), agents)
}
