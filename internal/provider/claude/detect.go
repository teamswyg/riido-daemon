package claude

import (
	"context"

	"github.com/teamswyg/riido-contracts/hostintegration"
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
			HealthStatus:   hostintegration.ProviderHealthUnavailable,
			DiagnosticCode: hostintegration.ProviderDiagnosticExecutableMissing,
			Reason:         "claude executable not found on PATH and " + EnvOverride + " is not set",
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
		HealthStatus:      hostintegration.ProviderHealthHealthy,
		DiagnosticCode:    hostintegration.ProviderDiagnosticNone,
	}
	switch claudeAuthProbe(ctx, exe) {
	case detectutil.AuthProbeAuthenticated:
	case detectutil.AuthProbeUnauthenticated:
		res.Available = false
		res.HealthStatus = hostintegration.ProviderHealthUnavailable
		res.DiagnosticCode = hostintegration.ProviderDiagnosticLoginRequired
		res.Reason = claudeAuthRecoveryMessage
	case detectutil.AuthProbeUnknown:
		res.Available = false
		res.HealthStatus = hostintegration.ProviderHealthUnknown
		res.DiagnosticCode = hostintegration.ProviderDiagnosticAuthProbeFailed
		res.Reason = "provider authentication probe did not complete"
	}
	if v, reason := detectutil.CLIVersionProbe(ctx, exe); reason == "" {
		res.Version = v
		res.Metadata["raw_version"] = v
	} else {
		res.Metadata["version_probe"] = reason
		if res.HealthStatus == hostintegration.ProviderHealthHealthy {
			res.HealthStatus = hostintegration.ProviderHealthUnknown
			res.DiagnosticCode = hostintegration.ProviderDiagnosticVersionProbeFailed
			res.Reason = reason
		}
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
