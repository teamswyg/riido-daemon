// Package cursor owns the C4 run-scope adapter for the Cursor Agent CLI.
//
// As of 2026-05 the cursor-agent CLI exposes `-p`, `--output-format`,
// `--yolo`, `--workspace`, and `--trust` at the root level. The
// historical `chat` subcommand is no longer accepted on current builds:
// cursor-agent treats the literal token "chat" as prompt text.
//
// To stay robust across cursor-agent versions, this adapter exposes
// three explicit launch profiles:
//
//   - ProfileRootPrint        cursor-agent -p <prompt> --output-format stream-json --workspace <cwd> --trust
//   - ProfileAgentSubcommand  cursor-agent agent -p ... --workspace <cwd> --trust (some builds require it)
//   - ProfileLegacyChat       cursor-agent chat -p ... --workspace <cwd> --trust (legacy; opt-in only)
//
// Default is ProfileRootPrint. ProfileLegacyChat MUST NOT be selected
// unless Detect (still deferred — Step 5) has confirmed that the local
// cursor-agent recognizes `chat` as a subcommand.
//
// --yolo and unsupported features (system prompt, max turns) follow the
// same explicit-opt-in / surface-as-Warning discipline as before.
package cursor

import (
	"fmt"
	"strconv"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

const Name = "cursor"
const DefaultExecutable = "cursor-agent"

func BlockedArgs() []string {
	return providercap.ProtocolCriticalArgs(providercap.ProtocolCursorAgentStreamJSON)
}

// Profile selects which cursor-agent CLI launch shape to use.
type Profile string

const (
	// ProfileRootPrint is the current cursor-agent CLI shape (2026-05+).
	// Pass -p / --output-format at the root, no subcommand.
	ProfileRootPrint Profile = "root-print"
	// ProfileAgentSubcommand uses `cursor-agent agent -p ...` for builds
	// that require the `agent` subcommand.
	ProfileAgentSubcommand Profile = "agent-subcommand"
	// ProfileLegacyChat uses `cursor-agent chat -p ...`. Opt-in only —
	// current cursor-agent treats `chat` as prompt text. Detect must
	// confirm the local build accepts this shape before selecting it.
	ProfileLegacyChat Profile = "legacy-chat"
)

// DefaultProfile is the launch shape used when StartOptions.Profile is
// empty. Until Detect can probe the local cursor-agent, the safe
// default is the current root-print form.
const DefaultProfile = ProfileRootPrint

type StartOptions struct {
	Executable string
	// Profile picks the cursor-agent launch shape. Empty → DefaultProfile.
	Profile Profile
	// AllowYolo opts into Cursor's --yolo (auto-approve every tool).
	// Default false. Must be selected by the caller's security policy.
	// The matrix that decides whether --yolo is permissible for the
	// current trust tier lives in docs/20-domain/security.md §5
	// (ExposesUnsafePermissionBypass) and is resolved by
	// internal/policy.EvaluateUnsafeBypass.
	AllowYolo bool
	// TrustTier and UnsafeBypassAllowed are consulted only when AllowYolo
	// is true. Host / Unknown deny regardless of UnsafeBypassAllowed;
	// isolated tiers also require bundle allow.
	TrustTier           policy.TrustTier
	UnsafeBypassAllowed bool
}

func BuildStart(req agentbridge.StartRequest, opts StartOptions) (agentbridge.StartCommand, error) {
	exe := opts.Executable
	if exe == "" {
		exe = DefaultExecutable
	}
	profile := opts.Profile
	if profile == "" {
		profile = DefaultProfile
	}

	var args []string
	switch profile {
	case ProfileRootPrint:
		args = []string{
			"-p", req.Prompt,
			"--output-format", "stream-json",
		}
	case ProfileAgentSubcommand:
		args = []string{
			"agent",
			"-p", req.Prompt,
			"--output-format", "stream-json",
		}
	case ProfileLegacyChat:
		args = []string{
			"chat",
			"-p", req.Prompt,
			"--output-format", "stream-json",
		}
	default:
		return agentbridge.StartCommand{}, fmt.Errorf("cursor: unknown profile %q (allowed: root-print, agent-subcommand, legacy-chat)", profile)
	}

	if opts.AllowYolo {
		decision := policy.EvaluateUnsafeBypass(policy.UnsafeBypassInput{
			TrustTier:    opts.TrustTier,
			Surface:      policy.UnsafeBypassCursorYolo,
			BundleAllows: opts.UnsafeBypassAllowed,
		})
		if !decision.Allowed {
			return agentbridge.StartCommand{}, fmt.Errorf("cursor: %s: %s", decision.Code, decision.Reason)
		}
		args = append(args, "--yolo")
	}
	if req.Cwd != "" {
		args = append(args, "--workspace", req.Cwd)
		// Headless Cursor Agent refuses to run in an untrusted workspace.
		// The daemon supplies a task-scoped workdir, so this acknowledges
		// that selected workspace without enabling Cursor's unsafe --yolo
		// auto-approval surface.
		args = append(args, "--trust")
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	if req.ResumeSessionID != "" {
		args = append(args, "--resume", req.ResumeSessionID)
	}

	kept, dropped := agentbridge.FilterBlockedArgs(req.CustomArgs, BlockedArgs())
	args = append(args, kept...)

	// Cursor doesn't support these; surface as warnings.
	if req.SystemPrompt != "" {
		dropped = append(dropped, "unsupported:system_prompt")
	}
	if req.MaxTurns > 0 {
		dropped = append(dropped, "unsupported:max_turns="+strconv.Itoa(req.MaxTurns))
	}

	env := make([]string, 0, len(req.Env))
	for k, v := range req.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return agentbridge.StartCommand{
		Executable:  exe,
		Args:        args,
		Env:         env,
		Dir:         req.Cwd,
		StdinMode:   agentbridge.StdinNone,
		DroppedArgs: dropped,
	}, nil
}
