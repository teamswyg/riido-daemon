package codex

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRPCActorRegisterAndResolve(t *testing.T) {
	a := StartRPCActor(context.Background())
	defer a.Close()

	id := a.NextID()
	resultCh := a.Register(id)
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

	a.Resolve(99999, map[string]any{"x": 1}, nil)
}
