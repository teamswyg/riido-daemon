package claude

import "github.com/teamswyg/riido-daemon/internal/policy"

// StartOptions carries Claude-specific knobs the daemon's run policy hands to
// BuildStart in addition to the provider-neutral agentbridge.StartRequest.
type StartOptions struct {
	// Executable overrides the binary path. Falls back to DefaultExecutable.
	Executable string
	// PermissionMode is REQUIRED; there is no default.
	PermissionMode PermissionMode
	// TrustTier and UnsafeBypassAllowed are consulted only for bypassPermissions.
	TrustTier           policy.TrustTier
	UnsafeBypassAllowed bool
	// MCPConfigPath enables strict MCP config with the exact config file path.
	MCPConfigPath string
	// PermissionPromptToolName routes Claude permission prompts to an MCP tool.
	PermissionPromptToolName string
}
