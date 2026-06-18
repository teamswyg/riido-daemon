package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func registerCodexRuntime(t *testing.T, plane *Plane) {
	t.Helper()
	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "runtime-1",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.codex.available": true,
		},
		DeviceName: "mac-mini",
	}); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}
}
