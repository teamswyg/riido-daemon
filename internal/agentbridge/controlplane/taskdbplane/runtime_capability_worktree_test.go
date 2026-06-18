package taskdbplane

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestRuntimeCapabilityForProviderReadsWorktreeSurface(t *testing.T) {
	capability, ok := runtimeCapabilityForProvider(
		openclawRuntimeCapability(false),
		"openclaw",
	)
	if !ok {
		t.Fatal("expected provider capability")
	}
	if capability.SupportsWorktree {
		t.Fatalf("worktree support must mirror runtime registry, got %+v", capability)
	}
	if !capability.SupportsStreaming || !capability.SupportsResume || !capability.SupportsUsage {
		t.Fatalf("other support flags not preserved: %+v", capability)
	}
}

func openclawRuntimeCapability(supportsWorktree bool) controlplane.RegisteredRuntime {
	prefix := "provider.openclaw."
	return controlplane.RegisteredRuntime{
		RuntimeRegistration: controlplane.RuntimeRegistration{
			RuntimeID: "runtime-1",
			Capabilities: map[string]bool{
				prefix + "available":                    true,
				prefix + "requires_experimental_opt_in": true,
				prefix + "supports_streaming":           true,
				prefix + "supports_resume":              true,
				prefix + "supports_usage":               true,
				prefix + "supports_worktree":            supportsWorktree,
			},
			CapabilityAttributes: map[string]string{
				prefix + "compatibility_status":   "experimental",
				prefix + "capability_fingerprint": "fp-openclaw",
			},
		},
	}
}
