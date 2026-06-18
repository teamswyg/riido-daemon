package runtimeactor

import "testing"

func assertClaudeExecutionFingerprint(t *testing.T, capability Capability) {
	t.Helper()
	if capability.AdapterID != "claude" {
		t.Fatalf("adapter id missing: %+v", capability)
	}
	if capability.AdapterVersion != "riido-agentbridge-adapter.v1" {
		t.Fatalf("adapter version missing: %+v", capability)
	}
	if capability.ProtocolVersion != "v1" || capability.CapabilityFingerprint == "" {
		t.Fatalf("fingerprint fields missing: %+v", capability)
	}
}

func assertClaudeSurfaceFlags(t *testing.T, capability Capability) {
	t.Helper()
	if !capability.SupportsStreaming || !capability.SupportsResume || !capability.SupportsSystem ||
		!capability.SupportsMaxTurns || !capability.SupportsMCP || !capability.SupportsToolHooks ||
		!capability.SupportsUsage || !capability.SupportsWorktree {
		t.Fatalf("surface flags were not preserved: %+v", capability)
	}
	if capability.SupportsFileEvents {
		t.Fatalf("file events must stay false until structured file events exist: %+v", capability)
	}
}
