package controlplane

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestMemorySourceClaimReturnsNoTaskWhenEmpty(t *testing.T) {
	src := NewMemorySource()
	req, err := src.ClaimTask(context.Background(), "rt-1")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if req != nil {
		t.Fatalf("expected nil request, got %+v", req)
	}
}

func TestMemorySourceQueueAndClaim(t *testing.T) {
	src := NewMemorySource()
	src.Enqueue(bridge.TaskRequest{ID: "t-1", Provider: "claude", Prompt: "hi"})
	src.Enqueue(bridge.TaskRequest{ID: "t-2", Provider: "codex", Prompt: "yo"})

	first, _ := src.ClaimTask(context.Background(), "rt-1")
	second, _ := src.ClaimTask(context.Background(), "rt-1")
	third, _ := src.ClaimTask(context.Background(), "rt-1")

	if first == nil || first.ID != "t-1" {
		t.Fatalf("first: %+v", first)
	}
	if second == nil || second.ID != "t-2" {
		t.Fatalf("second: %+v", second)
	}
	if third != nil {
		t.Fatalf("expected empty queue, got %+v", third)
	}
}

func TestMemorySourceRegisterDeregisterHeartbeat(t *testing.T) {
	src := newHeartbeatMemorySource()
	registerMemoryRuntime(t, src)
	assertMemoryRuntimeRegistered(t, src)
	heartbeatMemoryRuntime(t, src)
	deregisterMemoryRuntime(t, src)
}

func TestMemorySourceWatchCancellation(t *testing.T) {
	src := NewMemorySource()
	src.Enqueue(bridge.TaskRequest{ID: "t-1", Provider: "claude"})
	_, _ = src.ClaimTask(context.Background(), "rt-1")

	ch, err := src.WatchCancellation(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	src.Cancel("t-1", errors.New("user cancel"))
	select {
	case cause := <-ch:
		if cause == nil || cause.Error() != "user cancel" {
			t.Fatalf("cause: %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation not delivered")
	}
}
