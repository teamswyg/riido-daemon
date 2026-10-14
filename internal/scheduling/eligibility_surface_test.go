package scheduling

import (
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

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
	assertReason(t, got, "MISSING_REQUIRED_SURFACE", SurfaceSystemPrompt)
}

func TestEvaluateCapabilityMissingRequiredWorktreeSurface(t *testing.T) {
	got := EvaluateCapability(TaskRequirements{
		Provider:         "openclaw",
		RequiredSurfaces: []RequiredSurface{SurfaceWorktree},
	}, RuntimeCapability{
		Provider:            "openclaw",
		Available:           true,
		CompatibilityStatus: capability.CompatExperimental,
		SupportsWorktree:    false,
	})
	assertReason(t, got, "MISSING_REQUIRED_SURFACE", SurfaceWorktree)
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
	assertReason(t, got, "UNKNOWN_REQUIRED_SURFACE", "")
}

func assertReason(t *testing.T, got Eligibility, code string, surface RequiredSurface) {
	t.Helper()
	if got.Eligible {
		t.Fatalf("expected ineligible")
	}
	if got.Reasons[0].Code != code {
		t.Fatalf("reason: %+v", got.Reasons)
	}
	if surface != "" && got.Reasons[0].Surface != surface {
		t.Fatalf("reason: %+v", got.Reasons)
	}
}
