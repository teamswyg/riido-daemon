package saasplane

import (
	"context"
	"testing"
)

func TestPlaneRemovesCancelWatcherWhenContextEnds(t *testing.T) {
	run := newCancelWatcherPlane(t)

	ctx, cancel := context.WithCancel(t.Context())
	ch := watchCancellation(t, run.plane, ctx, "asn-ctx-done", "ctx done")
	assertCancelWatcherCount(t, run.plane, 1, "")

	cancel()

	assertCancelWatcherClosed(t, ch, "cancel watcher channel should close on context end")
	eventuallyCancelWatcherCount(t, run.plane, 0)
}

func TestPlaneContextCleanupDoesNotCloseReplacementCancelWatcher(t *testing.T) {
	run := newCancelWatcherPlane(t)

	firstCtx, firstCancel := context.WithCancel(t.Context())
	firstCh := watchCancellation(t, run.plane, firstCtx, "asn-replaced", "first")

	secondCtx, secondCancel := context.WithCancel(t.Context())
	defer secondCancel()
	secondCh := watchCancellation(t, run.plane, secondCtx, "asn-replaced", "second")

	assertCancelWatcherClosed(t, firstCh, "replaced cancel watcher channel should close")

	firstCancel()
	waitForStaleCancelWatcherCleanup()
	assertCancelWatcherCount(t, run.plane, 1, "after stale cleanup")
	assertCancelWatcherOpen(t, secondCh, "stale context cleanup closed the replacement watcher")
}
