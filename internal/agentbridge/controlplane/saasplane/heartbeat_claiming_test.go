package saasplane

import (
	"context"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneDeliversCancellationFromUnrefreshedHeartbeat(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	first := assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "first",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	}
	fake.enqueue(first)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask first: %v", err)
	}
	if req == nil || req.Metadata[MetadataAssignmentID] != first.ID {
		t.Fatalf("first claim = %+v", req)
	}
	if err := plane.StartTask(context.Background(), req.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	cancelCh, err := plane.WatchCancellation(context.Background(), req.ID)
	if err != nil {
		t.Fatalf("WatchCancellation: %v", err)
	}

	fake.staleHeartbeatIDs[first.ID] = true
	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      RuntimeIDForAgent("daemon-1", AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}),
		RunningTaskIDs: []string{req.ID},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	select {
	case cause := <-cancelCh:
		if cause == nil || !strings.Contains(cause.Error(), "heartbeat lease stale") {
			t.Fatalf("cancel cause = %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for stale heartbeat cancellation")
	}
}

func TestPlaneClaimsActiveAssignmentAfterLocalStateLoss(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	active := assignmentcontract.Assignment{
		ID:                       "asn-active",
		TaskID:                   "task-active",
		ComponentID:              "component-1",
		AgentID:                  "jykim1",
		RuntimeProvider:          "codex",
		Prompt:                   "resume active assignment",
		State:                    assignmentcontract.AssignmentLeased,
		LeaseToken:               "lease-active",
		AllowExperimentalRuntime: true,
		ResumeSessionID:          "sess-initial",
		ProviderSessionID:        "sess-current",
	}
	fake.activeNext(active.AgentID, active)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask active: %v", err)
	}
	if req == nil || req.ID != active.ID || req.Metadata[MetadataAssignmentID] != active.ID {
		t.Fatalf("active claim = %+v", req)
	}
	if !req.AllowExperimentalRuntime {
		t.Fatal("active assignment should preserve experimental opt-in")
	}
	if req.ResumeSessionID != active.ProviderSessionID {
		t.Fatalf("active assignment resume_session_id = %q, want provider session %q", req.ResumeSessionID, active.ProviderSessionID)
	}
}

func TestPlaneFailsActiveAssignmentWithoutSessionAfterLocalStateLoss(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	active := assignmentcontract.Assignment{
		ID:              "asn-active",
		TaskID:          "task-active",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "would duplicate side effects if started fresh",
		State:           assignmentcontract.AssignmentRunning,
		LeaseToken:      "lease-active",
	}
	fake.activeNext(active.AgentID, active)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask active without session: %v", err)
	}
	if req != nil {
		t.Fatalf("active assignment without session must not be fresh-started: %+v", req)
	}
	if len(fake.events) != 1 {
		t.Fatalf("events = %+v, want one recovery failure event", fake.events)
	}
	event := fake.events[0]
	if event.EventType != assignmentcontract.EventAssignmentFailed ||
		event.State != assignmentcontract.AssignmentFailed ||
		event.Metadata[recoveryMetadataKey] != recoveryFreshStartCode {
		t.Fatalf("recovery failure event = %+v", event)
	}
	if !strings.Contains(event.Message, "refusing fresh start") {
		t.Fatalf("recovery failure message = %q", event.Message)
	}
}

func TestPlanePollsOnlyRuntimeScopedAgent(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-2",
		TaskID:          "task-b",
		ComponentID:     "component-1",
		AgentID:         "jykim2",
		RuntimeProvider: "codex",
		Prompt:          "second agent task",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-2",
	})
	agents := []AgentBinding{
		{AgentID: "jykim1", RuntimeProvider: "codex"},
		{AgentID: "jykim2", RuntimeProvider: "codex"},
	}
	plane := newTestPlane(t, fake.URL(), agents)
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), RuntimeIDForAgent("daemon-1", agents[0]))
	if err != nil {
		t.Fatalf("ClaimTask jykim1: %v", err)
	}
	if req != nil {
		t.Fatalf("jykim1 runtime claimed another agent task: %+v", req)
	}
	req, err = plane.ClaimTask(context.Background(), RuntimeIDForAgent("daemon-1", agents[1]))
	if err != nil {
		t.Fatalf("ClaimTask jykim2: %v", err)
	}
	if req == nil || req.ID != "asn-2" || req.Metadata[MetadataAgentID] != "jykim2" {
		t.Fatalf("jykim2 claim = %+v", req)
	}
}

func TestPlaneSendsBearerToken(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.bearerToken = "secret"
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane := newTestPlaneWithToken(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}}, "secret")
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask with token: %v", err)
	}
	if req == nil || req.ID != "asn-1" {
		t.Fatalf("request = %+v", req)
	}
}
