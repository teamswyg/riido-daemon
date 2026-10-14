package codex

import (
	"context"
	"testing"
)

func TestRPCActorAssignsMonotonicIDs(t *testing.T) {
	a := StartRPCActor(context.Background())
	defer a.Close()

	id1 := a.NextID()
	id2 := a.NextID()
	id3 := a.NextID()
	if id2 != id1+1 || id3 != id2+1 {
		t.Fatalf("ids not monotonic: %d %d %d", id1, id2, id3)
	}
}
