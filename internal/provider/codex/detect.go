package codex

import (
	"context"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

const EnvOverride = "RIIDO_CODEX_PATH"

func Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	exe, ok := detectutil.ResolveExecutable(DefaultExecutable, envValue(env, EnvOverride))
	if !ok {
		return agentbridge.DetectResult{
			HealthStatus:   hostintegration.ProviderHealthUnavailable,
			DiagnosticCode: hostintegration.ProviderDiagnosticExecutableMissing,
			Reason:         "codex executable not found on PATH and " + EnvOverride + " is not set",
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
		HealthStatus:      hostintegration.ProviderHealthHealthy,
		DiagnosticCode:    hostintegration.ProviderDiagnosticNone,
	}
	switch detectutil.AuthStatusProbe(ctx, exe, "login", "status") {
	case detectutil.AuthProbeAuthenticated:
	case detectutil.AuthProbeUnauthenticated:
		res.Available = false
		res.HealthStatus = hostintegration.ProviderHealthUnavailable
		res.DiagnosticCode = hostintegration.ProviderDiagnosticLoginRequired
		res.Reason = "provider login is required"
	case detectutil.AuthProbeUnknown:
		res.HealthStatus = hostintegration.ProviderHealthUnknown
		res.DiagnosticCode = hostintegration.ProviderDiagnosticAuthProbeFailed
		res.Reason = "provider authentication probe did not complete"
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
