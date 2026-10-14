package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorUsesAssignmentIDAsExecutionKey(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		MaxConcurrent: 2,
	})

	if _, err := a.Submit(context.Background(), assignmentIdentityTaskRequest("asn-1")); err != nil {
		t.Fatalf("submit first assignment: %v", err)
	}
	if _, err := a.Submit(context.Background(), assignmentIdentityTaskRequest("asn-2")); err != nil {
		t.Fatalf("submit second assignment: %v", err)
	}
	r1 := waitForRunning(t, p, 0, time.Second)
	r2 := waitForRunning(t, p, 1, time.Second)

	assertRunningAssignmentIDs(t, a, "asn-1", "asn-2")
	assertHeartbeatAssignmentIDs(t, a, "asn-1", "asn-2")
	assertLogicalTaskCancelRejected(t, a)

	if err := a.Cancel(context.Background(), "asn-2", "stop second assignment"); err != nil {
		t.Fatalf("cancel second assignment: %v", err)
	}
	assertAssignmentKilled(t, r2)
	r1.EmitExit(0, nil)
}
