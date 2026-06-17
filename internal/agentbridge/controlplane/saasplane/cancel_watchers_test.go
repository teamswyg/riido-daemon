package saasplane

import (
	"context"
	"testing"
	"time"
)

func TestPlaneRemovesCancelWatcherWhenContextEnds(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	ctx, cancel := context.WithCancel(t.Context())
	ch, err := plane.WatchCancellation(ctx, "asn-ctx-done")
	if err != nil {
		t.Fatalf("WatchCancellation: %v", err)
	}
	if got := cancelWatcherCount(t, plane); got != 1 {
		t.Fatalf("cancel watcher count = %d, want 1", got)
	}

	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("cancel watcher channel should close on context end")
		}
	case <-time.After(time.Second):
		t.Fatal("cancel watcher channel was not closed after context end")
	}
	eventuallyCancelWatcherCount(t, plane, 0)
}

func TestPlaneContextCleanupDoesNotCloseReplacementCancelWatcher(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newTestPlane(t, fake.URL(), []AgentBinding{{AgentID: "jykim1", RuntimeProvider: "codex"}})
	defer plane.Close()

	firstCtx, firstCancel := context.WithCancel(t.Context())
	firstCh, err := plane.WatchCancellation(firstCtx, "asn-replaced")
	if err != nil {
		t.Fatalf("WatchCancellation first: %v", err)
	}

	secondCtx, secondCancel := context.WithCancel(t.Context())
	defer secondCancel()
	secondCh, err := plane.WatchCancellation(secondCtx, "asn-replaced")
	if err != nil {
		t.Fatalf("WatchCancellation second: %v", err)
	}

	select {
	case _, ok := <-firstCh:
		if ok {
			t.Fatal("replaced cancel watcher channel should close")
		}
	case <-time.After(time.Second):
		t.Fatal("replaced cancel watcher channel was not closed")
	}

	firstCancel()
	time.Sleep(50 * time.Millisecond)
	if got := cancelWatcherCount(t, plane); got != 1 {
		t.Fatalf("cancel watcher count after stale cleanup = %d, want 1", got)
	}
	select {
	case _, ok := <-secondCh:
		if !ok {
			t.Fatal("stale context cleanup closed the replacement watcher")
		}
	default:
	}
}

func cancelWatcherCount(t *testing.T, plane *Plane) int {
	t.Helper()
	var count int
	if err := plane.withState(context.Background(), func(s *planeState) {
		count = len(s.cancelWatchers)
	}); err != nil {
		t.Fatalf("read cancel watcher count: %v", err)
	}
	return count
}

func eventuallyCancelWatcherCount(t *testing.T, plane *Plane, want int) {
	t.Helper()
	deadline := time.After(time.Second)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if got := cancelWatcherCount(t, plane); got == want {
			return
		}
		select {
		case <-tick.C:
		case <-deadline:
			t.Fatalf("cancel watcher count = %d, want %d", cancelWatcherCount(t, plane), want)
		}
	}
}
