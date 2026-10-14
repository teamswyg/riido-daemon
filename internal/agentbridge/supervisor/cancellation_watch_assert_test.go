package supervisor

import (
	"context"
	"testing"
	"time"
)

func expectCancellationWatchContext(t *testing.T, source *cancelSource) context.Context {
	t.Helper()
	select {
	case watchCtx := <-source.watchCtxs:
		return watchCtx
	case <-time.After(supervisorCancellationTestTimeout):
		t.Fatal("cancellation watcher was not started")
		return context.Background()
	}
}

func expectContextDone(t *testing.T, ctx context.Context, msg string) {
	t.Helper()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal(msg)
	}
}
