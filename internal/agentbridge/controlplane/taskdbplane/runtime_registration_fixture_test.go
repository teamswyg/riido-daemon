package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func registerRuntimeForProvider(
	t *testing.T,
	plane *Plane,
	runtimeID string,
	provider string,
	slotLimit int,
	slotsInUse int,
) {
	t.Helper()

	prefix := "provider." + provider + "."
	err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		RuntimeID:  runtimeID,
		Provider:   "multi",
		SlotLimit:  slotLimit,
		SlotsInUse: slotsInUse,
		Capabilities: map[string]bool{
			prefix + "available":                    true,
			prefix + "supports_streaming":           true,
			prefix + "supports_resume":              true,
			prefix + "supports_system":              true,
			prefix + "supports_max_turns":           true,
			prefix + "supports_mcp":                 true,
			prefix + "supports_tool_hooks":          true,
			prefix + "supports_usage":               true,
			prefix + "supports_worktree":            true,
			prefix + "requires_experimental_opt_in": false,
		},
		CapabilityAttributes: map[string]string{
			prefix + "compatibility_status":   "supported",
			prefix + "capability_fingerprint": runtimeID + "-fp",
		},
	})
	if err != nil {
		t.Fatalf("RegisterRuntime %s: %v", runtimeID, err)
	}
}
