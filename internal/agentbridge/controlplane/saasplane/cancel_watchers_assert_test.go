package saasplane

import (
	"context"
	"testing"
	"time"
)

func assertCancelWatcherClosed(t *testing.T, ch <-chan error, msg string) {
	t.Helper()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal(msg)
		}
	case <-time.After(time.Second):
		t.Fatalf("%s: channel was not closed", msg)
	}
}

func assertCancelWatcherOpen(t *testing.T, ch <-chan error, msg string) {
	t.Helper()
	select {
	case _, ok := <-ch:
		if !ok {
			t.Fatal(msg)
		}
	default:
	}
}

func assertCancelWatcherCount(t *testing.T, plane *Plane, want int, suffix string) {
	t.Helper()
	if got := cancelWatcherCount(t, plane); got != want {
		t.Fatalf("cancel watcher count %s= %d, want %d", suffix, got, want)
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
