package main

import (
	"maps"
	"os"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

func providerDetectEnvironment(env agentbridge.DetectEnv) agentbridge.DetectEnv {
	values := make(map[string]string, len(env.EnvOverride)+4)
	maps.Copy(values, env.EnvOverride)
	for _, key := range []string{claude.EnvOverride, codex.EnvOverride, openclaw.EnvOverride, cursor.EnvOverride} {
		if _, exists := values[key]; exists {
			continue
		}
		if value, exists := os.LookupEnv(key); exists {
			values[key] = value
		}
	}
	env.EnvOverride = values
	return env
}
