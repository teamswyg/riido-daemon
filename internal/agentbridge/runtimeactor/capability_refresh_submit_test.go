package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorRefreshesUnavailableCapabilityAfterTTL(t *testing.T) {
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	provider := lateProvider("late")
	actor, process := startActor(t, Config{
		Adapters:               []agentbridge.Adapter{provider},
		CapabilityRefreshEvery: time.Second,
		Now:                    func() time.Time { return now },
	})

	provider.detected = lateAvailableDetectResult("1.2.3", "late")
	submitTask := func(id string) error {
		_, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: id, Provider: "late"})
		return err
	}
	if err := submitTask("before-ttl"); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("submit before ttl should use cached unavailable capability, got %v", err)
	}
	if process.count() != 0 {
		t.Fatalf("submit before ttl should not spawn provider, got %d", process.count())
	}

	now = now.Add(2 * time.Second)
	if err := submitTask("after-ttl"); err != nil {
		t.Fatalf("submit after ttl should refresh provider capability: %v", err)
	}
	if process.count() != 1 {
		t.Fatalf("submit after ttl should spawn provider once, got %d", process.count())
	}
	assertSingleAvailableCapability(t, actorStatusCapabilities(t, actor), "1.2.3")
}
