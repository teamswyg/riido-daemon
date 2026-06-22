package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

func resolveProviderExecutable(provider provider, override string) (string, bool) {
	if provider.ID == "openclaw" {
		return resolveOpenClawExecutable(override)
	}
	return resolveExecutable(provider.DefaultExecutable, override)
}

func resolveOpenClawExecutable(override string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), providerVersionTimeout)
	defer cancel()
	res, err := openclaw.Detect(ctx, agentbridge.DetectEnv{
		EnvOverride: map[string]string{openclaw.EnvOverride: override},
	})
	if err != nil {
		return resolveExecutable(openclaw.DefaultExecutable, override)
	}
	if !res.Available {
		return res.Executable, false
	}
	return res.Executable, true
}
