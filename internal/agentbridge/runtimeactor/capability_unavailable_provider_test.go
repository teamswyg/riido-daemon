package runtimeactor

import (
	"context"
	"testing"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorReconcilesUnavailableProviderAsBlocked(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "cursor", detected: agentbridge.DetectResult{
				Available: false,
				Reason:    "cursor-agent missing",
			}},
		},
	})

	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	capability := status.Capabilities[0]
	if capability.Available {
		t.Fatalf("capability should be unavailable: %+v", capability)
	}
	if capability.CompatibilityStatus != string(providercap.CompatBlocked) {
		t.Fatalf("unavailable provider must be blocked: %+v", capability)
	}
	if capability.ProtocolKind != string(providercap.ProtocolCursorAgentStreamJSON) {
		t.Fatalf("cursor protocol kind missing: %+v", capability)
	}
}
