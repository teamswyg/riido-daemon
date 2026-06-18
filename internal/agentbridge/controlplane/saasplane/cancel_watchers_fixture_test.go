package saasplane

import (
	"context"
	"testing"
	"time"
)

type cancelWatcherPlaneRun struct {
	plane *Plane
}

func newCancelWatcherPlane(t *testing.T) cancelWatcherPlaneRun {
	t.Helper()
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	t.Cleanup(plane.Close)
	return cancelWatcherPlaneRun{plane: plane}
}

func watchCancellation(
	t *testing.T,
	plane *Plane,
	ctx context.Context,
	executionID string,
	label string,
) <-chan error {
	t.Helper()
	ch, err := plane.WatchCancellation(ctx, executionID)
	if err != nil {
		t.Fatalf("WatchCancellation %s: %v", label, err)
	}
	return ch
}

func waitForStaleCancelWatcherCleanup() {
	time.Sleep(50 * time.Millisecond)
}
