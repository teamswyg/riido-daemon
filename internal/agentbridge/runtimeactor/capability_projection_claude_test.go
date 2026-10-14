package runtimeactor

import (
	"testing"
	"time"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestRuntimeActorReconcilesDetectResultToProviderCapability(t *testing.T) {
	fixedNow := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	actor, _ := startActor(t, Config{
		RuntimeID:           "rt-cap",
		PolicyBundleVersion: "policy-bundle.test.v1",
		Adapters:            []agentbridge.Adapter{claudeLikeAdapter()},
		Now:                 func() time.Time { return fixedNow },
	})

	caps := actorStatusCapabilities(t, actor)
	if len(caps) != 1 {
		t.Fatalf("capabilities: %+v", caps)
	}
	capability := caps[0]
	if capability.ProtocolKind != string(providercap.ProtocolClaudeStreamJSON) {
		t.Fatalf("protocol kind: %+v", capability)
	}
	if capability.CompatibilityStatus != string(providercap.CompatSupported) {
		t.Fatalf("compatibility status: %+v", capability)
	}
	assertClaudeExecutionFingerprint(t, capability)
	assertClaudeSurfaceFlags(t, capability)
}

func claudeLikeAdapter() *stubAdapter {
	return &stubAdapter{name: "claude", detected: agentbridge.DetectResult{
		Available:         true,
		Version:           "2.1.150",
		Executable:        "/usr/local/bin/claude",
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
		SupportsToolHooks: true,
		SupportsUsage:     true,
	}}
}
