package runtimeactor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorStatusRefreshesUnavailableCapabilityAfterTTL(t *testing.T) {
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	provider := lateProvider("late-status")
	actor, _ := startActor(t, Config{
		Adapters:               []agentbridge.Adapter{provider},
		CapabilityRefreshEvery: time.Second,
		Now:                    func() time.Time { return now },
	})

	provider.detected = lateAvailableDetectResult("2.0.0", "late-status")
	caps := actorStatusCapabilities(t, actor)
	if len(caps) != 1 || caps[0].Available {
		t.Fatalf("status before ttl should keep cached unavailable capability: %+v", caps)
	}

	now = now.Add(2 * time.Second)
	assertSingleAvailableCapability(t, actorStatusCapabilities(t, actor), "2.0.0")
}

func lateProvider(name string) *stubAdapter {
	return &stubAdapter{
		name:     name,
		detected: agentbridge.DetectResult{Available: false, Reason: "not installed"},
	}
}
