package cursor

import (
	"context"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

// Detect resolves the cursor-agent executable and reports capability metadata.
func Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	exe, ok := detectutil.ResolveExecutable(DefaultExecutable, envValue(env, EnvOverride))
	if !ok {
		return agentbridge.DetectResult{
			HealthStatus:   hostintegration.ProviderHealthUnavailable,
			DiagnosticCode: hostintegration.ProviderDiagnosticExecutableMissing,
			Reason:         "cursor-agent executable not found on PATH and " + EnvOverride + " is not set",
		}, nil
	}
	res := detectedCursor(exe)
	switch detectutil.AuthStatusProbe(ctx, exe, "status") {
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
		HealthStatus:      hostintegration.ProviderHealthHealthy,
		DiagnosticCode:    hostintegration.ProviderDiagnosticNone,
	}
}
