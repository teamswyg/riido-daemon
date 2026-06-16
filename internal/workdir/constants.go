package workdir

const (
	// NativeConfigVersionSchemaVersion is owned by docs/20-domain/workspace.md §6.
	NativeConfigVersionSchemaVersion = 1

	// ArchiveRecordSchemaVersion is owned by docs/20-domain/workspace.md §3.2.
	ArchiveRecordSchemaVersion = "riido-workdir-archive.v1"

	// NativeConfigHookModeInstructionOnly records that no provider-native hook
	// script/settings file has been materialized yet; enforcement currently
	// comes from the primary instruction file.
	NativeConfigHookModeInstructionOnly = "instruction-only"

	// NativeConfigHookModeClaudeCommandHooks records that Claude Code command
	// hooks were materialized into the per-task workdir.
	NativeConfigHookModeClaudeCommandHooks = "claude-command-hooks"

	// NativeConfigHomeModeDisabled records that C7 denied provider-native
	// config-home materialization for this run. The primary instruction file
	// remains, but provider settings files under ConfigHomeDir are stripped.
	NativeConfigHomeModeDisabled = "config-home-disabled"

	// RetentionModeKeepInPlace is the local daemon default: mark the run
	// archived without deleting the workdir tree.
	RetentionModeKeepInPlace = "keep-in-place"
)
