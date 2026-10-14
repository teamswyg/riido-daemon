package runtimeactor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorDetectsCapabilitiesOnStart(t *testing.T) {
	available := &stubAdapter{name: "available", detected: availableDetectResult("1.0")}
	missing := &stubAdapter{
		name:     "missing",
		detected: agentbridge.DetectResult{Available: false, Reason: "not installed"},
	}
	actor, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{available, missing},
	})

	caps := actorStatusCapabilities(t, actor)
	if len(caps) != 2 {
		t.Fatalf("want 2 capabilities, got %d: %+v", len(caps), caps)
	}
	byProvider := capabilitiesByProvider(caps)
	if !byProvider["available"].Available || byProvider["available"].Version != "1.0" {
		t.Fatalf("available capability: %+v", byProvider["available"])
	}
	if byProvider["missing"].Available || byProvider["missing"].Reason != "not installed" {
		t.Fatalf("missing capability: %+v", byProvider["missing"])
	}
}

func availableDetectResult(version string) agentbridge.DetectResult {
	return agentbridge.DetectResult{
		Available:  true,
		Version:    version,
		Executable: "/usr/bin/available",
	}
}
