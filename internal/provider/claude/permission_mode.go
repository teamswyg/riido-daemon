package claude

// PermissionMode is Claude's tool-permission policy. There is no default.
type PermissionMode string

const (
	// PermissionModeApproval maps Riido's safe approval mode to Claude's
	// `default` permission mode.
	PermissionModeApproval PermissionMode = "default"
	// PermissionModeAcceptEdits auto-approves edit/write tools but still gates
	// bash/shell tools.
	PermissionModeAcceptEdits PermissionMode = "acceptEdits"
	// PermissionModePlan maps to Claude's read-only exploration mode.
	PermissionModePlan PermissionMode = "plan"
	// PermissionModeBypassDangerous maps to Anthropic's bypassPermissions mode.
	PermissionModeBypassDangerous PermissionMode = "bypassPermissions"
)
