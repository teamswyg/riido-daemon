package scheduling

import (
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestSelectRuntimeChoosesEligibleRuntimeByCapability(t *testing.T) {
	selection, ok := SelectRuntime(TaskRequirements{
		Provider:         "codex",
		RequiredSurfaces: []RequiredSurface{SurfaceStructuredEventStream, SurfaceUsage},
	}, []RuntimeCapability{
		{
			RuntimeID:             "rt-unavailable",
			Provider:              "codex",
			CapabilityFingerprint: "fp-unavailable",
			Available:             false,
			CompatibilityStatus:   capability.CompatSupported,
			SupportsStreaming:     true,
			SupportsUsage:         true,
		},
		{
			RuntimeID:             "rt-best",
			Provider:              "codex",
			CapabilityFingerprint: "fp-best",
			SlotLimit:             2,
			SlotsInUse:            1,
			Available:             true,
			CompatibilityStatus:   capability.CompatSupported,
			SupportsStreaming:     true,
			SupportsUsage:         true,
		},
	})
	if !ok {
		t.Fatalf("expected selection, got %+v", selection)
	}
	if selection.Runtime.RuntimeID != "rt-best" || selection.Eligibility.RuntimeID != "rt-best" {
		t.Fatalf("selection = %+v", selection)
	}
	if len(selection.Rejected) != 1 || selection.Rejected[0].RuntimeID != "rt-unavailable" {
		t.Fatalf("rejected = %+v", selection.Rejected)
	}
}

func TestSelectRuntimePrefersSupportedThenHeadroom(t *testing.T) {
	req := TaskRequirements{Provider: "claude", AllowExperimentalRuntime: true}
	candidates := []RuntimeCapability{
		{
			RuntimeID:                 "rt-experimental-empty",
			Provider:                  "claude",
			CapabilityFingerprint:     "fp-exp",
			SlotLimit:                 10,
			SlotsInUse:                0,
			Available:                 true,
			CompatibilityStatus:       capability.CompatExperimental,
			RequiresExperimentalOptIn: true,
		},
		{
			RuntimeID:             "rt-supported-busy",
			Provider:              "claude",
			CapabilityFingerprint: "fp-supported-busy",
			SlotLimit:             2,
			SlotsInUse:            1,
			Available:             true,
			CompatibilityStatus:   capability.CompatSupported,
		},
		{
			RuntimeID:             "rt-supported-empty",
			Provider:              "claude",
			CapabilityFingerprint: "fp-supported-empty",
			SlotLimit:             2,
			SlotsInUse:            0,
			Available:             true,
			CompatibilityStatus:   capability.CompatSupported,
		},
	}
	selection, ok := SelectRuntime(req, candidates)
	if !ok {
		t.Fatal("expected eligible runtime")
	}
	if selection.Runtime.RuntimeID != "rt-supported-empty" {
		t.Fatalf("selected runtime = %+v", selection.Runtime)
	}
	if candidates[0].RuntimeID != "rt-experimental-empty" {
		t.Fatalf("SelectRuntime must not mutate candidate order: %+v", candidates)
	}
}

func TestSelectRuntimeRejectsSlotExhaustedPool(t *testing.T) {
	selection, ok := SelectRuntime(TaskRequirements{Provider: "cursor"}, []RuntimeCapability{
		{
			RuntimeID:             "rt-full",
			Provider:              "cursor",
			CapabilityFingerprint: "fp-full",
			SlotLimit:             1,
			SlotsInUse:            1,
			Available:             true,
			CompatibilityStatus:   capability.CompatSupported,
		},
	})
	if ok {
		t.Fatalf("expected no selected runtime, got %+v", selection)
	}
	if len(selection.Rejected) != 1 || selection.Rejected[0].Reasons[0].Code != "SLOT_EXHAUSTED" {
		t.Fatalf("rejected = %+v", selection.Rejected)
	}
}

func TestSelectRuntimeReturnsStableRejections(t *testing.T) {
	selection, ok := SelectRuntime(TaskRequirements{Provider: "openclaw"}, []RuntimeCapability{
		{RuntimeID: "rt-z", Provider: "codex", Available: true, CompatibilityStatus: capability.CompatSupported},
		{RuntimeID: "rt-a", Provider: "claude", Available: true, CompatibilityStatus: capability.CompatSupported},
	})
	if ok {
		t.Fatalf("expected no selection, got %+v", selection)
	}
	if len(selection.Rejected) != 2 || selection.Rejected[0].RuntimeID != "rt-a" || selection.Rejected[1].RuntimeID != "rt-z" {
		t.Fatalf("rejections not stable: %+v", selection.Rejected)
	}
}
