package saasplane

import (
	"context"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneKeepsSameTaskAssignmentsIndependent(t *testing.T) {
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
	second := first
	second.ID = "asn-2"
	second.Prompt = "second"
	second.LeaseToken = "lease-2"
	fake.enqueue(first)
	fake.enqueue(second)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req1, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask first: %v", err)
	}
	req2, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask second: %v", err)
	}
	if req1 == nil || req2 == nil || req1.ID != first.ID || req2.ID != second.ID {
		t.Fatalf("claims = %+v / %+v", req1, req2)
	}
	if req1.Metadata[controlplane.MetadataTaskID] != first.TaskID || req2.Metadata[controlplane.MetadataTaskID] != second.TaskID {
		t.Fatalf("logical task metadata lost: %+v / %+v", req1.Metadata, req2.Metadata)
	}
	if err := plane.StartTask(context.Background(), req1.ID); err != nil {
		t.Fatalf("StartTask first: %v", err)
	}
	if err := plane.StartTask(context.Background(), req2.ID); err != nil {
		t.Fatalf("StartTask second: %v", err)
	}
	cancel1, err := plane.WatchCancellation(context.Background(), req1.ID)
	if err != nil {
		t.Fatalf("WatchCancellation first: %v", err)
	}
	cancel2, err := plane.WatchCancellation(context.Background(), req2.ID)
	if err != nil {
		t.Fatalf("WatchCancellation second: %v", err)
	}

	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      RuntimeIDForAgent("daemon-1", AgentBinding{AgentID: "jykim1", RuntimeProvider: "codex"}),
		RunningTaskIDs: []string{req1.ID, req2.ID},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	heartbeats := fake.heartbeatsFor("jykim1")
	if len(heartbeats) != 1 || strings.Join(heartbeats[0].ActiveAssignmentIDs, ",") != "asn-1,asn-2" {
		t.Fatalf("heartbeats = %+v", heartbeats)
	}
	if err := plane.ReportEvent(context.Background(), req1.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "first progress"}); err != nil {
		t.Fatalf("ReportEvent first: %v", err)
	}
	if err := plane.ReportEvent(context.Background(), req2.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "second progress"}); err != nil {
		t.Fatalf("ReportEvent second: %v", err)
	}
	if last := fake.events[len(fake.events)-1]; last.AssignmentID != second.ID || last.TaskID != second.TaskID {
		t.Fatalf("second event identity = %+v", last)
	}

	fake.cancelNext(second.AgentID, second)
	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask cancel poll: %v", err)
	}
	if req != nil {
		t.Fatalf("cancel poll should not claim new task: %+v", req)
	}
	select {
	case cause := <-cancel2:
		if cause == nil || !strings.Contains(cause.Error(), second.ID) {
			t.Fatalf("second cancel cause = %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for second cancellation")
	}
	if _, ok := <-cancel2; ok {
		t.Fatal("second cancellation watcher should close after cancel")
	}
	select {
	case cause, ok := <-cancel1:
		t.Fatalf("first watcher should remain independent, cause=%v ok=%v", cause, ok)
	case <-time.After(20 * time.Millisecond):
	}
}

func TestPlaneClosesCancellationWatcherOnComplete(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentcontract.Assignment{
		ID:              "asn-1",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "complete",
		State:           assignmentcontract.AssignmentQueued,
		LeaseToken:      "lease-1",
	})
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	cancelCh, err := plane.WatchCancellation(context.Background(), req.ID)
	if err != nil {
		t.Fatalf("WatchCancellation: %v", err)
	}
	if err := plane.CompleteTask(context.Background(), req.ID, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "ok"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}
	select {
	case _, ok := <-cancelCh:
		if ok {
			t.Fatal("completion should close watcher without cancellation cause")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for watcher close")
	}
}
