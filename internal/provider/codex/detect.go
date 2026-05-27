package codex

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

const EnvOverride = "RIIDO_CODEX_PATH"

func Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	exe, ok := detectutil.ResolveExecutable(DefaultExecutable, envValue(env, EnvOverride))
	if !ok {
		return agentbridge.DetectResult{
			Available: false,
			Reason:    "codex executable not found on PATH and " + EnvOverride + " is not set",
		}, nil
	}
	res := agentbridge.DetectResult{
		Available:         true,
		Executable:        exe,
		SupportsStreaming: true,
		SupportsResume:    true,  // thread/resume in app-server
		SupportsSystem:    true,  // developer instructions
		SupportsMaxTurns:  false, // adapter-dependent; conservative
		SupportsMCP:       false, // partial / unsupported on app-server flavor
		SupportsToolHooks: true,
		SupportsUsage:     true,
		Metadata:          map[string]string{},
	}
	if v, ok := detectutil.VersionProbe(ctx, exe, "--version"); ok {
		res.Version = v
		res.Metadata["raw_version"] = v
	}
	return res, nil
}

func envValue(env agentbridge.DetectEnv, key string) string {
	if env.EnvOverride != nil {
		if v, ok := env.EnvOverride[key]; ok {
			return v
		}
	}
	return ""
}
