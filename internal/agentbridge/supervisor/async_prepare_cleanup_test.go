package supervisor

import (
	"context"
	"testing"
	"time"
)

func stopSupervisorAfterPrepare(
	t *testing.T,
	actor *Actor,
	cloneStarted <-chan struct{},
	cloneDone <-chan struct{},
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = actor.Stop(ctx)

	select {
	case <-cloneStarted:
		select {
		case <-cloneDone:
		case <-time.After(time.Second):
			t.Error("workspace materialization goroutine did not stop")
		}
	default:
	}
}
