// Package claude owns the C4 run-scope adapter for Anthropic's Claude Code CLI.
// It owns command construction, executable detection, stream-json parsing,
// translation, and stdin protocol framing. The adapter is a translator; it does
// NOT own a state machine of its own. agentbridge does.
//
// This package provides:
//   - The blocked-args list (protocol-critical flags the adapter sets
//     itself; custom args containing these are dropped).
//   - BuildStart: turns an agentbridge.StartRequest into a StartCommand.
//   - An explicit, required PermissionMode parameter. There is no default that
//     maps to bypassPermissions. See docs/20-domain/security.md.
//   - Detect/NewParser/Translate/NewProtocolDriver adapter hooks.
package claude

import (
	"fmt"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

// Name is the canonical adapter identifier.
const Name = "claude"

// DefaultExecutable is the binary name resolved on $PATH when no
// explicit executable is configured.
const DefaultExecutable = "claude"

// BlockedArgs lists the protocol-critical flags this adapter manages
// itself. Custom args containing any of these are dropped with a
// Warning event so the caller knows their override was ignored.
//
// --permission-mode is blocked here because the adapter must NOT allow
// callers to silently flip into bypassPermissions through free-form
// custom args. The mode is selected only via the explicit
// StartOptions.PermissionMode parameter, which is governed by the
// task's security policy
// (docs/20-domain/security.md).
func BlockedArgs() []string {
	return providercap.ProtocolCriticalArgs(providercap.ProtocolClaudeStreamJSON)
}

// PermissionMode is Claude's tool-permission policy. There is no
// default — callers must pick one explicitly.
type PermissionMode string

const (
	// PermissionModeApproval requires explicit approval for every tool
	// invocation. Safe default for unverified workspaces. Maps to
	// Claude's `--permission-mode default` (the prompt-on-every-tool
	// mode). Claude's actual flag values are: default / acceptEdits /
	// auto / bypassPermissions / dontAsk / plan — there is no literal
	// "approval" value, so we map our semantic name onto Claude's
	// `default`. Verified against `claude --version 2.1.150`.
	PermissionModeApproval PermissionMode = "default"
	// PermissionModeAcceptEdits auto-approves edit/write tools but
	// still gates bash/shell tools. Intermediate trust tier.
	PermissionModeAcceptEdits PermissionMode = "acceptEdits"
	// PermissionModePlan corresponds to Claude's `plan` mode (read-only
	// exploration; the agent forms a plan but doesn't apply changes).
	PermissionModePlan PermissionMode = "plan"
	// PermissionModeBypassDangerous == Anthropic's bypassPermissions /
	// --dangerously-skip-permissions. Container/VM-only.
	// Documented as REJECTED as a default in
	// docs/20-domain/security.md.
	PermissionModeBypassDangerous PermissionMode = "bypassPermissions"
)

// StartOptions carries Claude-specific knobs the daemon's run policy
// hands to BuildStart in addition to the provider-neutral
// agentbridge.StartRequest.
type StartOptions struct {
	// Executable overrides the binary path. Falls back to DefaultExecutable.
	Executable string
	// PermissionMode is REQUIRED — there is no default. The caller's
	// security policy must select one explicitly.
	PermissionMode PermissionMode
	// TrustTier and UnsafeBypassAllowed are consulted only when
	// PermissionMode is bypassPermissions. Host / Unknown deny regardless
	// of UnsafeBypassAllowed; isolated tiers also require bundle allow.
	TrustTier           policy.TrustTier
	UnsafeBypassAllowed bool
	// MCPConfigPath is the path to a serialized MCP JSON config. When
	// non-empty, both --strict-mcp-config and --mcp-config are set,
	// avoiding the strict-without-config trap where Claude sees strict MCP mode
	// but no config file path.
	MCPConfigPath string
}

// BuildStart turns an agentbridge.StartRequest + Claude-specific options
// into a StartCommand. Custom args from req are filtered against
// BlockedArgs; dropped args land in StartCommand.DroppedArgs so the
// session actor can emit a Warning event per spec §9.1.
func BuildStart(req agentbridge.StartRequest, opts StartOptions) (agentbridge.StartCommand, error) {
	if opts.PermissionMode == "" {
		return agentbridge.StartCommand{}, fmt.Errorf("%s: PermissionMode is required (no implicit bypass — see docs/20-domain/security.md)", Name)
	}
	if opts.PermissionMode == PermissionModeBypassDangerous {
		decision := policy.EvaluateUnsafeBypass(policy.UnsafeBypassInput{
			TrustTier:    opts.TrustTier,
			Surface:      policy.UnsafeBypassClaudePermissions,
			BundleAllows: opts.UnsafeBypassAllowed,
		})
		if !decision.Allowed {
			return agentbridge.StartCommand{}, fmt.Errorf("%s: %s: %s", Name, decision.Code, decision.Reason)
		}
	}
	exe := opts.Executable
	if exe == "" {
		exe = req.Executable
	}
	if exe == "" {
		exe = DefaultExecutable
	}

	args := []string{
		"-p",
		"--output-format", "stream-json",
		"--input-format", "stream-json",
		"--verbose",
		"--permission-mode", string(opts.PermissionMode),
	}

	tempFiles := []string{}
	if opts.MCPConfigPath != "" {
		args = append(args, "--strict-mcp-config", "--mcp-config", opts.MCPConfigPath)
		tempFiles = append(tempFiles, opts.MCPConfigPath)
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	if req.SystemPrompt != "" {
		args = append(args, "--append-system-prompt", req.SystemPrompt)
	}
	if req.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", req.MaxTurns))
	}
	if req.ResumeSessionID != "" {
		args = append(args, "--resume", req.ResumeSessionID)
	}

	kept, dropped := agentbridge.FilterBlockedArgs(req.CustomArgs, BlockedArgs())
	args = append(args, kept...)

	env := make([]string, 0, len(req.Env))
	for k, v := range req.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return agentbridge.StartCommand{
		Executable:  exe,
		Args:        args,
		Env:         env,
		Dir:         req.Cwd,
		StdinMode:   agentbridge.StdinPipe,
		DroppedArgs: dropped,
		TempFiles:   tempFiles,
	}, nil
}
