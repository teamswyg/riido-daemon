package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestCapabilityRefreshRejectsStaleGenerationAndEnvironment(t *testing.T) {
	a := &Actor{cfg: Config{Now: time.Now}}
	r := newCapabilityRefresh(1)
	reply := make(chan submitReply, 1)
	r.probes["fake"] = &capabilityProbe{generation: 2, fingerprint: "old", cancel: func() {}, pending: []*submitMsg{{req: bridge.TaskRequest{ID: "queued"}, reply: reply}}}
	r.pendingIDs["queued"] = true
	caps := []Capability{{Provider: "fake", CapabilityFingerprint: "new"}}
	times := map[string]time.Time{}
	if got := r.finish(a, capabilityProbeResult{provider: "fake", generation: 1, fingerprint: "old"}, caps, times, nil); got != nil {
		t.Fatal("old generation released submissions")
	}
	if r.probes["fake"] == nil {
		t.Fatal("old generation removed current probe")
	}
	if got := r.finish(a, capabilityProbeResult{provider: "fake", generation: 2, fingerprint: "old"}, caps, times, nil); got != nil {
		t.Fatal("old environment released submissions")
	}
	if err := (<-reply).err; !errors.Is(err, ErrRuntimePinViolated) {
		t.Fatalf("stale environment error=%v", err)
	}
	if caps[0].CapabilityFingerprint != "new" {
		t.Fatal("stale observation replaced current identity")
	}
}

func TestCancelledQueuedSubmitNeverStartsAfterProbe(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	adapter := &delayedRefreshAdapter{stubAdapter: stubAdapter{name: "fake", detected: fakeDetect("1.0.0")}, entered: make(chan struct{}), release: make(chan struct{})}
	actor, proc := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, CapabilityRefreshEvery: time.Second, Now: func() time.Time { return now }})
	now = now.Add(2 * time.Second)
	returned := make(chan error, 1)
	go func() {
		_, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "queued", Provider: "fake"})
		returned <- err
	}()
	select {
	case <-adapter.entered:
	case <-time.After(time.Second):
		t.Fatal("probe missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := actor.Cancel(ctx, "queued", "cancel"); err != nil {
		t.Fatal(err)
	}
	if err := <-returned; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error=%v", err)
	}
	close(adapter.release)
	awaitCapability(t, actor, func(c Capability) bool { return c.Available })
	if proc.count() != 0 {
		t.Fatal("cancelled queued submission spawned a process")
	}
}

func TestQueuedSubmitStopRefusesNewExecution(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	adapter := &delayedRefreshAdapter{stubAdapter: stubAdapter{name: "fake", detected: fakeDetect("1.0.0")}, entered: make(chan struct{}), release: make(chan struct{})}
	actor, proc := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, CapabilityRefreshEvery: time.Second, Now: func() time.Time { return now }})
	t.Cleanup(func() { close(adapter.release) })
	now = now.Add(2 * time.Second)
	returned := make(chan error, 1)
	go func() {
		_, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "queued", Provider: "fake"})
		returned <- err
	}()
	select {
	case <-adapter.entered:
	case <-time.After(time.Second):
		t.Fatal("probe missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := actor.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-returned; !errors.Is(err, ErrActorStopped) {
		t.Fatalf("stop error=%v", err)
	}
	if proc.count() != 0 {
		t.Fatal("stopped actor dispatched a queued submission")
	}
}
