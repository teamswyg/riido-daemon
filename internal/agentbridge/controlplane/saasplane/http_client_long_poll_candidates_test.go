package saasplane

import (
	"context"
	"testing"
	"time"
)

func TestPlaneShortPollsAllCandidatesThenLongPollsOne(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane, err := New(Config{
		BaseURL:  fake.URL(),
		DaemonID: "daemon-1",
		DeviceID: "device-1",
		Agents: []AgentBinding{
			{AgentID: "agent-a", RuntimeProvider: "codex"},
			{AgentID: "agent-b", RuntimeProvider: "codex"},
		},
		LongPollWait: 2500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	if req != nil {
		t.Fatalf("empty fake queue should not claim task: %+v", req)
	}
	agentA := fake.pollRequestsFor("agent-a")
	agentB := fake.pollRequestsFor("agent-b")
	if len(agentA) != 2 || len(agentB) != 1 {
		t.Fatalf("poll requests agent-a=%+v agent-b=%+v", agentA, agentB)
	}
	if agentA[0].WaitMs != 0 || agentB[0].WaitMs != 0 || agentA[1].WaitMs != 2500 {
		t.Fatalf("unexpected wait_ms distribution agent-a=%+v agent-b=%+v", agentA, agentB)
	}
}
