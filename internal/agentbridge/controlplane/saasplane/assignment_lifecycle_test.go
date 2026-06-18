package saasplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestPlaneKeepsSameTaskAssignmentsIndependent(t *testing.T) {
	fake, plane, first, second := setupSameTaskAssignments(t)
	defer plane.Close()

	req1, req2 := claimSameTaskAssignments(t, plane, first, second)
	startSameTaskAssignments(t, plane, req1.ID, req2.ID)
	cancel1, cancel2 := watchSameTaskAssignments(t, plane, req1.ID, req2.ID)

	reportSameTaskHeartbeat(t, fake, plane, req1.ID, req2.ID)
	reportSameTaskProgress(t, fake, plane, req1.ID, req2.ID, second)
	pollSecondAssignmentCancellation(t, fake, plane, second)
	assertSecondAssignmentCancelled(t, cancel2, second.ID)
	assertFirstAssignmentStillWatching(t, cancel1)
}

func TestPlaneClosesCancellationWatcherOnComplete(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.enqueue(assignmentLifecycleFixture("asn-1", "complete", "lease-1"))
	plane := newLifecyclePlane(t, fake)
	defer plane.Close()

	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}
	cancelCh, err := plane.WatchCancellation(context.Background(), req.ID)
	if err != nil {
		t.Fatalf("WatchCancellation: %v", err)
	}
	result := agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "ok"}
	if err := plane.CompleteTask(context.Background(), req.ID, result); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}
	assertWatcherClosedAfterCompletion(t, cancelCh)
}
