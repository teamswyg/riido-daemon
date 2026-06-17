package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestRuntimeActorUsesAssignmentIDAsExecutionKey(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		MaxConcurrent: 2,
	})
	req := func(id string) bridge.TaskRequest {
		return bridge.TaskRequest{
			ID:       id,
			Provider: "fake",
			Metadata: map[string]string{
				controlplane.MetadataTaskID: "task-a",
			},
		}
	}

	if _, err := a.Submit(context.Background(), req("asn-1")); err != nil {
		t.Fatalf("submit first assignment: %v", err)
	}
	if _, err := a.Submit(context.Background(), req("asn-2")); err != nil {
		t.Fatalf("submit second assignment: %v", err)
	}
	r1 := waitForRunning(t, p, 0, time.Second)
	r2 := waitForRunning(t, p, 1, time.Second)

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.RunningSessions != 2 {
		t.Fatalf("running sessions = %d, want 2", status.RunningSessions)
	}
	seen := map[string]bool{}
	for _, task := range status.RunningTasks {
		seen[task.TaskID] = true
	}
	if !seen["asn-1"] || !seen["asn-2"] || seen["task-a"] {
		t.Fatalf("running task ids = %+v, want assignment ids only", status.RunningTasks)
	}

	hb, err := a.HeartbeatPayload(context.Background())
	if err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	if got := hb.RunningTaskIDs; len(got) != 2 || got[0] != "asn-1" || got[1] != "asn-2" {
		t.Fatalf("heartbeat running ids = %v, want assignment ids", got)
	}
	if err := a.Cancel(context.Background(), "task-a", "logical id is not execution id"); !errors.Is(err, ErrUnknownTask) {
		t.Fatalf("cancel by logical task id = %v, want ErrUnknownTask", err)
	}

	if err := a.Cancel(context.Background(), "asn-2", "stop second assignment"); err != nil {
		t.Fatalf("cancel second assignment: %v", err)
	}
	select {
	case <-r2.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("second assignment process was not killed")
	}
	r1.EmitExit(0, nil)
}
