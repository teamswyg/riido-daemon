package saasplane

import (
	"context"
	"testing"
	"time"
)

func watchSameTaskAssignments(
	t *testing.T,
	plane *Plane,
	firstID string,
	secondID string,
) (<-chan error, <-chan error) {
	t.Helper()
	cancel1 := watchLifecycleCancellation(t, plane, firstID, "first")
	cancel2 := watchLifecycleCancellation(t, plane, secondID, "second")
	return cancel1, cancel2
}

func watchLifecycleCancellation(
	t *testing.T,
	plane *Plane,
	assignmentID string,
	label string,
) <-chan error {
	t.Helper()
	cancelCh, err := plane.WatchCancellation(context.Background(), assignmentID)
	if err != nil {
		t.Fatalf("WatchCancellation %s: %v", label, err)
	}
	return cancelCh
}

func assertWatcherClosedAfterCompletion(t *testing.T, cancelCh <-chan error) {
	t.Helper()
	select {
	case _, ok := <-cancelCh:
		if ok {
			t.Fatal("completion should close watcher without cancellation cause")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for watcher close")
	}
}
