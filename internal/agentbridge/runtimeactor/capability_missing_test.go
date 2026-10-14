package runtimeactor

import (
	"slices"
	"testing"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestMissingCapabilitiesIncludesProfileOwnedWorktree(t *testing.T) {
	missing := missingCapabilities(fullDetectResult(), openClawCapabilityProfile())
	if !slices.Contains(missing, providercap.CapabilityName("worktree")) {
		t.Fatalf("missing capabilities should expose worktree drift: %+v", missing)
	}
}

func TestMissingCapabilitiesOmitsSupportedWorktree(t *testing.T) {
	missing := missingCapabilities(fullDetectResult(), claudeCapabilityProfile())
	if slices.Contains(missing, providercap.CapabilityName("worktree")) {
		t.Fatalf("supported worktree should not be missing: %+v", missing)
	}
}

func fullDetectResult() agentbridge.DetectResult {
	return agentbridge.DetectResult{
		Available:         true,
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
		SupportsToolHooks: true,
		SupportsUsage:     true,
	}
}
