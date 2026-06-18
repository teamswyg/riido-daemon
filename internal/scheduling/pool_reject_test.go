package scheduling

import (
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

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
