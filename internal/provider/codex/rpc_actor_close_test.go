package codex

import (
	"context"
	"testing"
	"time"
)

func TestRPCActorCloseCancelsPending(t *testing.T) {
	a := StartRPCActor(context.Background())
	id := a.NextID()
	resultCh := a.Register(id)

	a.Close()

	select {
	case r := <-resultCh:
		if r.Err == nil {
			t.Fatal("expected error on actor close")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout: actor close did not release pending caller")
	}
}
