package runtimeactor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

type delayedRefreshAdapter struct {
	stubAdapter
	mu      sync.Mutex
	calls   int
	entered chan struct{}
	release chan struct{}
}

func (a *delayedRefreshAdapter) Detect(ctx context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	a.mu.Lock()
	a.calls++
	call := a.calls
	a.mu.Unlock()
	if call > 1 {
		if call == 2 {
			close(a.entered)
		}
		select {
		case <-a.release:
		case <-ctx.Done():
			return agentbridge.DetectResult{}, ctx.Err()
		}
	}
	return a.detected, nil
}

func TestSlowCapabilityRefreshKeepsStatusAndCancelResponsive(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	adapter := &delayedRefreshAdapter{stubAdapter: stubAdapter{name: "fake", detected: fakeDetect("1.0.0")}, entered: make(chan struct{}), release: make(chan struct{})}
	actor, proc := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, CapabilityRefreshEvery: time.Second, Now: func() time.Time { return now }})
	t.Cleanup(func() { close(adapter.release) })
	handle, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "running", Provider: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	_ = waitForRunning(t, proc, 0, time.Second)
	now = now.Add(2 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if _, err = actor.Status(ctx); err != nil {
		t.Fatalf("slow probe blocked Status: %v", err)
	}
	select {
	case <-adapter.entered:
	case <-ctx.Done():
		t.Fatal("refresh did not start")
	}
	if err = actor.Cancel(ctx, "running", "user cancelled"); err != nil {
		t.Fatalf("slow probe blocked Cancel: %v", err)
	}
	if res := expectTaskResult(t, handle.Result(), "cancel result"); res.Status != agentbridge.ResultCancelled {
		t.Fatalf("result=%+v", res)
	}
	if _, err = actor.Status(ctx); err != nil {
		t.Fatalf("in-flight probe blocked subsequent Status: %v", err)
	}
	adapter.mu.Lock()
	calls := adapter.calls
	adapter.mu.Unlock()
	if calls != 2 {
		t.Fatalf("concurrent status created %d Detect calls, want initial + single refresh", calls)
	}
}

func TestSlowCapabilityRefreshDoesNotBlockStop(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	adapter := &delayedRefreshAdapter{stubAdapter: stubAdapter{name: "fake", detected: fakeDetect("1.0.0")}, entered: make(chan struct{}), release: make(chan struct{})}
	actor, _ := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, CapabilityRefreshEvery: time.Second, Now: func() time.Time { return now }})
	t.Cleanup(func() { close(adapter.release) })
	now = now.Add(2 * time.Second)
	requestCtx, requestCancel := context.WithCancel(t.Context())
	defer requestCancel()
	go func() { _, _ = actor.Status(requestCtx) }()
	select {
	case <-adapter.entered:
	case <-time.After(time.Second):
		t.Fatal("refresh did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if err := actor.Stop(ctx); err != nil {
		t.Fatalf("slow Detect blocked Stop: %v", err)
	}
}
