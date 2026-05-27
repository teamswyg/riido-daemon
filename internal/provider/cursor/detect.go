package cursor

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

const EnvOverride = "RIIDO_CURSOR_PATH"

// Detect resolves the cursor-agent executable, reads --version, and
// inspects --help to pick a launch profile.
//
// The launch profile is reported in DetectResult.Metadata["profile"]
// (one of "root-print" / "agent-subcommand" / "legacy-chat"). The
// daemon (Step 4 RuntimeActor) reads this and threads it into
// StartOptions on each BuildStart call.
//
// Profile selection logic (conservative):
//   - If --help mentions "chat" as a subcommand line → "legacy-chat".
//   - Else if --help mentions an "agent" subcommand → "agent-subcommand".
//   - Else → "root-print" (safe default for current cursor-agent).
func Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	exe, ok := detectutil.ResolveExecutable(DefaultExecutable, envValue(env, EnvOverride))
	if !ok {
		return agentbridge.DetectResult{
			Available: false,
			Reason:    "cursor-agent executable not found on PATH and " + EnvOverride + " is not set",
		}, nil
	}
	res := agentbridge.DetectResult{
		Available:         true,
		Executable:        exe,
		SupportsStreaming: true,
		SupportsResume:    true,  // --resume
		SupportsSystem:    false, // no system prompt
		SupportsMaxTurns:  false, // no max turns
		SupportsMCP:       false,
		SupportsToolHooks: true,
		SupportsUsage:     true,
		Metadata:          map[string]string{},
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

func pickProfile(help string) Profile {
	lower := strings.ToLower(help)
	// Look for subcommand headers like "  chat" or "Commands:\n  chat".
	hasChatSubcommand := strings.Contains(lower, "\nchat") ||
		strings.Contains(lower, "  chat ") ||
		strings.Contains(lower, " chat  ")
	hasAgentSubcommand := strings.Contains(lower, "\nagent") ||
		strings.Contains(lower, "  agent ") ||
		strings.Contains(lower, " agent  ")

	switch {
	case hasChatSubcommand:
		return ProfileLegacyChat
	case hasAgentSubcommand:
		return ProfileAgentSubcommand
	default:
		return ProfileRootPrint
	}
}

func envValue(env agentbridge.DetectEnv, key string) string {
	if env.EnvOverride != nil {
		if v, ok := env.EnvOverride[key]; ok {
			return v
		}
	}
	return ""
}
