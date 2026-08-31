package openclaw

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func detectExecutable(ctx context.Context, exe string) agentbridge.DetectResult {
	base := agentbridge.DetectResult{
		Executable:        exe,
		SupportsStreaming: true,
		SupportsResume:    false, // agent exec owns an isolated run
		SupportsSystem:    false, // inlined into --message
		SupportsMaxTurns:  false,
		SupportsMCP:       false,
		SupportsToolHooks: false,
		SupportsUsage:     true,
		Metadata:          map[string]string{},
	}

	probe := detectutil.VersionProbeStrict(ctx, exe, "--version")
	if !probe.OK {
		base.Available = false
		base.Reason = "openclaw --version did not run to completion (timeout or signal); cannot enforce minimum version " + MinSupportedVersion
		return base
	}

	if probe.ExitCode != 0 {
		// Non-zero exit is authoritative: even if the output happens
		// to look like a version, refuse to lift it.
		base.Available = false
		base.Reason = sanitizeReason(probe.Output)
		// Leave Version empty — exit code says we have no trustworthy
		// version information.
		return base
	}

	parsed, ok := parseVersion(probe.Output)
	if !ok {
		base.Available = false
		base.Version = ""
		base.Reason = "openclaw --version output did not match the expected YYYY.M.D shape: " + sanitizeReason(probe.Output)
		return base
	}

	// Successful parse: record what we observed for diagnostics.
	base.Version = sanitizeReason(probe.Output)
	base.Metadata["raw_version"] = probe.Output

	minTuple, _ := parseVersion(MinSupportedVersion)
	if compareVersions(parsed, minTuple) < 0 {
		base.Available = false
		base.Reason = "openclaw " + base.Version + " is older than minimum supported " + MinSupportedVersion + " — upgrade openclaw"
		return base
	}
	help := detectutil.VersionProbeStrict(ctx, exe, "agent", "exec", "--help")
	if !help.OK || help.ExitCode != 0 || !strings.Contains(help.Output, "--json") {
		base.Available = false
		base.Reason = "openclaw agent exec --json is unavailable; upgrade openclaw"
		return base
	}

	base.Available = true
	base.Metadata["protocol"] = "agent-exec-json"
	auth := detectutil.VersionProbeStrict(ctx, exe, "models", "status", "--check", "--json")
	if !auth.OK || !json.Valid([]byte(auth.Output)) {
		base.Reason = "provider authentication probe did not complete"
		return base
	}
	switch auth.ExitCode {
	case 1:
		base.Available = false
		base.Reason = "provider login is required"
	case 2:
		base.Reason = "provider login requires attention"
	case 0:
	default:
		base.Reason = "provider authentication probe did not complete"
	}
	return base
}

func envValue(env agentbridge.DetectEnv, key string) string {
	if env.EnvOverride != nil {
		if v, ok := env.EnvOverride[key]; ok {
			return v
		}
	}
	return ""
}
