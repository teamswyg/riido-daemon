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

func TestEvaluateCapabilityMissingRequiredSurface(t *testing.T) {
	got := EvaluateCapability(TaskRequirements{
		Provider:         "cursor",
		RequiredSurfaces: []RequiredSurface{SurfaceSystemPrompt},
	}, RuntimeCapability{
		Provider:            "cursor",
		Available:           true,
		CompatibilityStatus: capability.CompatSupported,
		SupportsSystem:      false,
	})
	if got.Eligible {
		t.Fatalf("expected ineligible")
	}
	if got.Reasons[0].Code != "MISSING_REQUIRED_SURFACE" || got.Reasons[0].Surface != SurfaceSystemPrompt {
		t.Fatalf("reason: %+v", got.Reasons)
	}
}

func TestEvaluateCapabilityExperimentalRequiresOptIn(t *testing.T) {
	cap := RuntimeCapability{
		Provider:                  "codex",
		Available:                 true,
		CompatibilityStatus:       capability.CompatExperimental,
		RequiresExperimentalOptIn: true,
	}
	withoutOptIn := EvaluateCapability(TaskRequirements{Provider: "codex"}, cap)
	if withoutOptIn.Eligible {
		t.Fatal("experimental runtime must require explicit opt-in")
	}
	withOptIn := EvaluateCapability(TaskRequirements{Provider: "codex", AllowExperimentalRuntime: true}, cap)
	if !withOptIn.Eligible {
		t.Fatalf("opt-in should allow experimental runtime: %+v", withOptIn)
	}
}

func TestEvaluateCapabilityUnknownSurfaceFailsClosed(t *testing.T) {
	got := EvaluateCapability(TaskRequirements{
		Provider:         "claude",
		RequiredSurfaces: []RequiredSurface{"future_surface"},
	}, RuntimeCapability{
		Provider:            "claude",
		Available:           true,
		CompatibilityStatus: capability.CompatSupported,
	})
	if got.Eligible {
		t.Fatal("unknown required surface must fail closed")
	}
	if got.Reasons[0].Code != "UNKNOWN_REQUIRED_SURFACE" {
		t.Fatalf("reason: %+v", got.Reasons)
	}
}

func TestNormalizeRequiredSurfaces(t *testing.T) {
	got := NormalizeRequiredSurfaces([]RequiredSurface{" MCP ", "mcp", "system_prompt", ""})
	want := []RequiredSurface{SurfaceMCP, SurfaceSystemPrompt}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	}
}
