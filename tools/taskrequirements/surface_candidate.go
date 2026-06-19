package main

import (
	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
)

func candidateBase() scheduling.RuntimeCapability {
	return scheduling.RuntimeCapability{
		Provider:            "probe",
		Available:           true,
		CompatibilityStatus: capability.CompatSupported,
	}
}

func candidateWithSurface(name string, supported bool) (scheduling.RuntimeCapability, bool) {
	candidate := candidateBase()
	switch scheduling.RequiredSurface(name) {
	case scheduling.SurfaceStructuredEventStream:
		candidate.SupportsStreaming = supported
	case scheduling.SurfaceSessionResume:
		candidate.SupportsResume = supported
	case scheduling.SurfaceSystemPrompt:
		candidate.SupportsSystem = supported
	case scheduling.SurfaceMaxTurns:
		candidate.SupportsMaxTurns = supported
	case scheduling.SurfaceMCP:
		candidate.SupportsMCP = supported
	case scheduling.SurfaceToolHooks:
		candidate.SupportsToolHooks = supported
	case scheduling.SurfaceUsage:
		candidate.SupportsUsage = supported
	case scheduling.SurfaceWorktree:
		candidate.SupportsWorktree = supported
	default:
		return candidate, false
	}
	return candidate, true
}

func expectedCandidateField(name string) string {
	fields := map[string]string{
		"structured-event-stream": "SupportsStreaming",
		"session-resume":          "SupportsResume",
		"system-prompt":           "SupportsSystem",
		"max-turns":               "SupportsMaxTurns",
		"mcp":                     "SupportsMCP",
		"tool-hooks":              "SupportsToolHooks",
		"usage":                   "SupportsUsage",
		"worktree":                "SupportsWorktree",
	}
	return fields[name]
}
