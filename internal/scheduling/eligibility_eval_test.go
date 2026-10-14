package scheduling

import (
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestEvaluateCapabilityEligible(t *testing.T) {
	got := EvaluateCapability(TaskRequirements{
		Provider:                 "claude",
		RequiredSurfaces:         []RequiredSurface{SurfaceStructuredEventStream, SurfaceSystemPrompt},
		AllowExperimentalRuntime: false,
	}, RuntimeCapability{
		RuntimeID:             "rt-1",
		Provider:              "claude",
		CapabilityFingerprint: "fp-1",
		Available:             true,
		CompatibilityStatus:   capability.CompatSupported,
		SupportsStreaming:     true,
		SupportsSystem:        true,
	})
	if !got.Eligible {
		t.Fatalf("expected eligible, got %+v", got)
	}
	if got.RuntimeID != "rt-1" || got.CapabilityFingerprint != "fp-1" {
		t.Fatalf("runtime pin not preserved: %+v", got)
	}
}

func TestEvaluateCapabilityExperimentalRequiresOptIn(t *testing.T) {
	candidate := RuntimeCapability{
		Provider:                  "codex",
		Available:                 true,
		CompatibilityStatus:       capability.CompatExperimental,
		RequiresExperimentalOptIn: true,
	}
	withoutOptIn := EvaluateCapability(TaskRequirements{Provider: "codex"}, candidate)
	if withoutOptIn.Eligible {
		t.Fatal("experimental runtime must require explicit opt-in")
	}
	withOptIn := EvaluateCapability(TaskRequirements{Provider: "codex", AllowExperimentalRuntime: true}, candidate)
	if !withOptIn.Eligible {
		t.Fatalf("opt-in should allow experimental runtime: %+v", withOptIn)
	}
}
