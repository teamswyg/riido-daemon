package saasplane

import (
	"context"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func pollSecondAssignmentCancellation(
	t *testing.T,
	fake *fakeAssignmentServer,
	plane *Plane,
	second assignmentcontract.Assignment,
) {
	t.Helper()
	fake.cancelNext(second.AgentID, second)
	req, err := plane.ClaimTask(context.Background(), "daemon-1:codex")
	if err != nil {
		t.Fatalf("ClaimTask cancel poll: %v", err)
	}
	if req != nil {
		t.Fatalf("cancel poll should not claim new task: %+v", req)
	}
}

func assertSecondAssignmentCancelled(t *testing.T, cancelCh <-chan error, assignmentID string) {
	t.Helper()
	select {
	case cause := <-cancelCh:
		if cause == nil || !strings.Contains(cause.Error(), assignmentID) {
			t.Fatalf("second cancel cause = %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for second cancellation")
	}
	if _, ok := <-cancelCh; ok {
		t.Fatal("second cancellation watcher should close after cancel")
	}
}

func assertFirstAssignmentStillWatching(t *testing.T, cancelCh <-chan error) {
	t.Helper()
	select {
	case cause, ok := <-cancelCh:
		t.Fatalf("first watcher should remain independent, cause=%v ok=%v", cause, ok)
	case <-time.After(20 * time.Millisecond):
	}
}
