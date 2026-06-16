package saasplane

import "testing"

func TestRuntimeIDForAgentRoundTripsAgentAndProvider(t *testing.T) {
	runtimeID := RuntimeIDForAgent("daemon:one", AgentBinding{
		AgentID:         "agent one",
		RuntimeProvider: "Claude-Code",
	})

	agentID, ok := agentFromRuntimeID(runtimeID)
	if !ok || agentID != "agent one" {
		t.Fatalf("agentFromRuntimeID(%q) = %q, %v", runtimeID, agentID, ok)
	}
	if got := providerFromRuntimeID(runtimeID); got != "claude_code" {
		t.Fatalf("providerFromRuntimeID(%q) = %q, want claude_code", runtimeID, got)
	}
}

func TestProviderFromRuntimeIDFallsBackToLastSegment(t *testing.T) {
	if got := providerFromRuntimeID("daemon-1:codex"); got != "codex" {
		t.Fatalf("providerFromRuntimeID fallback = %q, want codex", got)
	}
}
