// Package workdir is the C6 Workspace adapter: per-task workdir trees
// and provider-native config file injection (CLAUDE.md / AGENTS.md /
// GEMINI.md / ...).
//
// What this package owns:
//   - Workspace tree layout:
//     <root>/<workspace>/tasks/<task>/runs/<run>/
//     {workdir,output,logs,artifacts,native-config,ir}/ and the
//     .gc_meta.json marker plus archive.json manifest used by lifecycle
//     retention.
//   - The generated provider→native-config file plan registry.
//   - The provider-native config manifest materialization evidence
//     written under .riido/native-config-manifest.json.
//   - The 4-section runtime-config template (Identity / CLI catalog /
//     Hard rules / Workflow) from spec §10 Phase 7.
//   - workspace_id enforcement: empty workspace IDs are rejected
//     (multica.md §6.1 "workspace_id 필수").
//
// What this package does NOT own:
//   - The C7 policy bundle that DECIDES what goes into
//     the rule set. workdir just renders what it is given.
//   - Retention TTL evaluation. The daemon decides when to archive or
//     clean up; workdir provides deterministic filesystem helpers.
package workdir

import (
	"time"
)

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

// TaskID is the per-task identity that determines the workdir tree path.
type TaskID struct {
	Workspace string
	Task      string
	Run       string
}

// Workspace is the result of Prepare: the on-disk tree paths.
type Workspace struct {
	Root         string // <root>/<workspace>/tasks/<task>/runs/<run>/
	Workdir      string // <root>/<workspace>/tasks/<task>/runs/<run>/workdir/
	Output       string // <root>/<workspace>/tasks/<task>/runs/<run>/output/
	Logs         string // <root>/<workspace>/tasks/<task>/runs/<run>/logs/
	Artifacts    string // <root>/<workspace>/tasks/<task>/runs/<run>/artifacts/
	NativeConfig string // <root>/<workspace>/tasks/<task>/runs/<run>/native-config/
	IR           string // <root>/<workspace>/tasks/<task>/runs/<run>/ir/
}

// RuntimeConfig is the inputs the workdir adapter renders into the
// provider's native config file (per spec §10 Phase 7, §7).
//
// The four sections mirror Multica's native-config 4단 structure
// (multica.md §7).
type RuntimeConfig struct {
	Provider                   string   // e.g. "claude", "codex"
	ProtocolKind               string   // C3 protocol selected for the run
	TelemetryContractPlacement string   // prompt/system-prompt/... when injected
	NativeHookMode             string   // C7 decision result; empty uses the provider plan default
	NativeConfigHomeMode       string   // C7 decision result; empty uses the provider plan default
	Identity                   string   // "You are: <agent name> (id: ...)"
	CLICatalog                 []string // command examples
	HardRules                  []string // invariants the agent must follow
	Workflow                   string   // workflow branch label (chat|quick-create|...)
}

// ProviderNativeConfigPlan is the deterministic file plan for one provider's
// native config materialization.
type ProviderNativeConfigPlan struct {
	ProviderKind           string
	PrimaryInstructionFile string
	ManifestFile           string
	HookMode               string
	ConfigHomeDir          string
	ProviderSettingsFiles  []string
	HookFiles              []string
	ExtraFiles             []string
}

// GeneratedFiles returns the relative file paths materialized by the plan.
func (p ProviderNativeConfigPlan) GeneratedFiles() []string {
	files := []string{p.PrimaryInstructionFile, p.ManifestFile}
	files = append(files, p.ProviderSettingsFiles...)
	files = append(files, p.HookFiles...)
	files = append(files, p.ExtraFiles...)
	return sortedUniquePaths(files)
}

// NativeConfigManifest is the replayable evidence of what C6 materialized
// into a run workdir and its native-config copy.
type NativeConfigManifest struct {
	SchemaVersion              string   `json:"schema_version"`
	ProviderKind               string   `json:"provider_kind"`
	ProtocolKind               string   `json:"protocol_kind,omitempty"`
	PrimaryInstructionFile     string   `json:"primary_instruction_file"`
	ManifestFile               string   `json:"manifest_file"`
	HookMode                   string   `json:"hook_mode"`
	ConfigHomeDir              string   `json:"config_home_dir,omitempty"`
	ProviderSettingsFiles      []string `json:"provider_settings_files,omitempty"`
	HookFiles                  []string `json:"hook_files,omitempty"`
	TelemetryContractPlacement string   `json:"telemetry_contract_placement,omitempty"`
	Workflow                   string   `json:"workflow"`
	GeneratedFiles             []string `json:"generated_files"`
}

// ArchiveRequest is the terminal-run input for Archive.
type ArchiveRequest struct {
	ResultStatus string
	ArchivedAt   time.Time
}

// ArchiveRecord is the local archive manifest written at run-root/archive.json.
type ArchiveRecord struct {
	SchemaVersion string    `json:"schema_version"`
	WorkdirPath   string    `json:"workdir_path"`
	ArchiveURI    string    `json:"archive_uri"`
	RetentionMode string    `json:"retention_mode"`
	ResultStatus  string    `json:"result_status"`
	ArchivedAt    time.Time `json:"archived_at"`
}

// CleanupRequest defines an explicit retention cleanup pass. The
// daemon supplies ArchivedBefore from its Factor 12 retention config.
type CleanupRequest struct {
	ArchivedBefore time.Time
	RemovedAt      time.Time
}

// CleanupRecord describes one run root removed by cleanup.
type CleanupRecord struct {
	RunRoot   string        `json:"run_root"`
	Archive   ArchiveRecord `json:"archive"`
	RemovedAt time.Time     `json:"removed_at"`
}

// CleanupResult summarizes an archived-run cleanup pass.
type CleanupResult struct {
	ScannedArchiveRecords int             `json:"scanned_archive_records"`
	Removed               []CleanupRecord `json:"removed"`
}

// NativeConfigVersionInput is the C6 execution-context fingerprint input.
type NativeConfigVersionInput struct {
	PolicyBundleVersion string
	ProviderKind        string
	ProtocolKind        string
}

type nativeConfigVersionDoc struct {
	PolicyBundleVersion string           `json:"policyBundleVersion"`
	NativeConfigPlan    nativeConfigPlan `json:"nativeConfigPlan"`
	SchemaVersion       int              `json:"schemaVersion"`
	Unknown             map[string]any   `json:"unknown,omitempty"`
}

type nativeConfigPlan struct {
	ProviderKind       string                 `json:"providerKind"`
	ProtocolKind       string                 `json:"protocolKind"`
	InjectedFiles      []nativeConfigFileHash `json:"injectedFiles"`
	HookScriptVersions []nativeConfigFileHash `json:"hookScriptVersions,omitempty"`
	WrapperManifestSHA string                 `json:"wrapperManifestSha,omitempty"`
}

type nativeConfigFileHash struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// Adapter is the port. The supervisor calls Prepare per claimed task and
// InjectRuntimeConfig before handing the workdir to the runtime actor.
type Adapter interface {
	Prepare(TaskID) (Workspace, error)
	InjectRuntimeConfig(Workspace, RuntimeConfig) error
}
