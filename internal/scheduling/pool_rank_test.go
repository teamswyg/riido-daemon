package scheduling

import (
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

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
