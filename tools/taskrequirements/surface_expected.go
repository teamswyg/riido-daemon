package main

func expectedCapabilityFlag(name string) string {
	flags := map[string]string{
		"structured-event-stream": "SupportsStructuredEventStream",
		"session-resume":          "SupportsResume",
		"system-prompt":           "SupportsSystemPrompt",
		"max-turns":               "SupportsMaxTurns",
		"mcp":                     "SupportsMCP",
		"tool-hooks":              "SupportsHookEvents",
		"usage":                   "SupportsUsageMetrics",
		"worktree":                "SupportsWorktree",
	}
	return flags[name]
}

func expectedSchedulingConstant(name string) string {
	constants := map[string]string{
		"structured-event-stream": "SurfaceStructuredEventStream",
		"session-resume":          "SurfaceSessionResume",
		"system-prompt":           "SurfaceSystemPrompt",
		"max-turns":               "SurfaceMaxTurns",
		"mcp":                     "SurfaceMCP",
		"tool-hooks":              "SurfaceToolHooks",
		"usage":                   "SurfaceUsage",
		"worktree":                "SurfaceWorktree",
	}
	return constants[name]
}
