package runtimeactor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func claudeCapabilityDetectResult(executable string) agentbridge.DetectResult {
	return agentbridge.DetectResult{
		Available:         true,
		Version:           "2.1.150",
		Executable:        executable,
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
		SupportsToolHooks: true,
		SupportsUsage:     true,
	}
}
