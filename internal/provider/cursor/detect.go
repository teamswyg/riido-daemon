package cursor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

// Detect resolves the cursor-agent executable and reports capability metadata.
func Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	exe, ok := detectutil.ResolveExecutable(DefaultExecutable, envValue(env, EnvOverride))
	if !ok {
		return agentbridge.DetectResult{
			Available: false,
			Reason:    "cursor-agent executable not found on PATH and " + EnvOverride + " is not set",
		}, nil
	}
	res := detectedCursor(exe)
	if v, ok := detectutil.VersionProbe(ctx, exe, "--version"); ok {
		res.Version = v
		res.Metadata["raw_version"] = v
	}
	if help, ok := detectutil.VersionProbe(ctx, exe, "--help"); ok {
		res.Metadata["profile"] = string(pickProfile(help))
	} else {
		res.Metadata["profile"] = string(ProfileRootPrint)
	}
	return res, nil
}

func detectedCursor(exe string) agentbridge.DetectResult {
	return agentbridge.DetectResult{
		Available:         true,
		Executable:        exe,
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    false,
		SupportsMaxTurns:  false,
		SupportsMCP:       false,
		SupportsToolHooks: true,
		SupportsUsage:     true,
		Metadata:          map[string]string{},
	}
}
