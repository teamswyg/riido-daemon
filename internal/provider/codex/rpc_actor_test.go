package codex

import (
	"context"
	"errors"
	"testing"
	"time"
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

func TestRPCActorRegisterAndResolve(t *testing.T) {
	a := StartRPCActor(context.Background())
	defer a.Close()

	id := a.NextID()
	resultCh := a.Register(id)

	// Resolving with the matching id delivers the response.
	a.Resolve(id, map[string]any{"thread_id": "t1"}, nil)

	select {
	case r := <-resultCh:
		if r.Err != nil {
			t.Fatalf("got error: %v", r.Err)
		}
		if r.Result["thread_id"] != "t1" {
			t.Fatalf("result: %+v", r.Result)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for resolve")
	}
}

func TestRPCActorResolveError(t *testing.T) {
	a := StartRPCActor(context.Background())
	defer a.Close()

	id := a.NextID()
	resultCh := a.Register(id)
	a.Resolve(id, nil, errors.New("rpc failed"))

	r := <-resultCh
	if r.Err == nil || r.Err.Error() != "rpc failed" {
		t.Fatalf("expected rpc failed, got %v", r.Err)
	}
}

func TestRPCActorResolveUnknownIDIsNoop(t *testing.T) {
	a := StartRPCActor(context.Background())
	defer a.Close()

	// Resolving an unknown id must not panic and must not block.
	a.Resolve(99999, map[string]any{"x": 1}, nil)
}

func TestRPCActorCloseCancelsPending(t *testing.T) {
	a := StartRPCActor(context.Background())
	id := a.NextID()
	resultCh := a.Register(id)

	a.Close()

	// Pending callers receive an error result (no leak, no block).
	select {
	case r := <-resultCh:
		if r.Err == nil {
			t.Fatal("expected error on actor close")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout — actor close did not release pending caller")
	}
}

// The actor must NOT use a mutex internally — invariant of the
// concurrency rule. We can't assert this from the test, but we assert
// behavior under -race in the rest of the package's tests.
