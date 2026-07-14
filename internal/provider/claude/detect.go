package claude

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

// EnvOverride is the env var callers may set to pin the claude
// executable when PATH lookup is unreliable (GUI-launched daemons).
const EnvOverride = "RIIDO_CLAUDE_PATH"

// Detect resolves the executable, verifies authentication, and reads
// --version. Authentication is required for runnable capability.
//
// When the binary is missing, returns Available=false with a clear
// Reason (NOT an error) so the daemon can surface it as a runtime
// capability gap.
func Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	exe, ok := detectutil.ResolveExecutable(DefaultExecutable, envValue(env, EnvOverride))
	if !ok {
		return agentbridge.DetectResult{
			Available: false,
			Reason:    "claude executable not found on PATH and " + EnvOverride + " is not set",
		}, nil
	}
	if !claudeAuthenticated(ctx, exe) {
		return agentbridge.DetectResult{
			Available:  false,
			Executable: exe,
			Reason:     claudeAuthRecoveryMessage,
		}, nil
	}
	res := agentbridge.DetectResult{
		Available:         true,
		Executable:        exe,
		SupportsStreaming: true,
		SupportsResume:    true,
		SupportsSystem:    true,
		SupportsMaxTurns:  true,
		SupportsMCP:       true,
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
