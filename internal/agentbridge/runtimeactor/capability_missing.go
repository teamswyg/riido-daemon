package runtimeactor

import (
	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func missingCapabilities(res agentbridge.DetectResult) []providercap.CapabilityName {
	checks := []struct {
		name providercap.CapabilityName
		ok   bool
	}{
		{"structured-event-stream", res.SupportsStreaming},
		{"session-resume", res.SupportsResume},
		{"system-prompt", res.SupportsSystem},
		{"max-turns", res.SupportsMaxTurns},
		{"mcp", res.SupportsMCP},
		{"tool-hooks", res.SupportsToolHooks},
		{"usage", res.SupportsUsage},
	}
	out := []providercap.CapabilityName{}
	for _, check := range checks {
		if !check.ok {
			out = append(out, check.name)
		}
	}
	return out
}
