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
