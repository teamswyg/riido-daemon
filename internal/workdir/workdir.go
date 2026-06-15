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
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
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

// Archiver is an optional port implemented by adapters that can record
// terminal workspace lifecycle state.
type Archiver interface {
	Archive(Workspace, ArchiveRequest) (ArchiveRecord, error)
}

// Cleaner is an optional port implemented by adapters that can delete
// archived run roots once an operator-supplied retention cutoff expires.
type Cleaner interface {
	CleanupArchivedBefore(context.Context, CleanupRequest) (CleanupResult, error)
}

// FSAdapter is the filesystem implementation rooted at a single path.
type FSAdapter struct {
	root string
}

// NewFSAdapter constructs an adapter rooted at root. The root is created
// lazily by Prepare.
func NewFSAdapter(root string) *FSAdapter { return &FSAdapter{root: root} }

// Prepare creates the per-task workspace tree and writes the GC marker.
// Returns an error when workspace id is empty (security/isolation gate).
func (a *FSAdapter) Prepare(id TaskID) (Workspace, error) {
	if strings.TrimSpace(id.Workspace) == "" {
		return Workspace{}, errors.New("workdir: workspace id is required")
	}
	if strings.TrimSpace(id.Task) == "" {
		return Workspace{}, errors.New("workdir: task id is required")
	}
	runID := strings.TrimSpace(id.Run)
	if runID == "" {
		runID = id.Task
	}
	if !safePathSegment(id.Workspace) || !safePathSegment(id.Task) || !safePathSegment(runID) {
		return Workspace{}, errors.New("workdir: workspace, task, or run id contains a path separator or traversal")
	}

	taskRoot := filepath.Join(a.root, id.Workspace, "tasks", id.Task, "runs", runID)
	ws := Workspace{
		Root:         taskRoot,
		Workdir:      filepath.Join(taskRoot, "workdir"),
		Output:       filepath.Join(taskRoot, "output"),
		Logs:         filepath.Join(taskRoot, "logs"),
		Artifacts:    filepath.Join(taskRoot, "artifacts"),
		NativeConfig: filepath.Join(taskRoot, "native-config"),
		IR:           filepath.Join(taskRoot, "ir"),
	}
	for _, dir := range []string{ws.Workdir, ws.Output, ws.Logs, ws.Artifacts, ws.NativeConfig, ws.IR} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Workspace{}, fmt.Errorf("workdir: mkdir %s: %w", dir, err)
		}
	}

	meta := map[string]any{
		"workspace_id": id.Workspace,
		"task_id":      id.Task,
		"run_id":       runID,
		"created_at":   time.Now().UTC().Format(time.RFC3339Nano),
	}
	metaBytes, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(ws.Root, ".gc_meta.json"), metaBytes, 0o644); err != nil {
		return Workspace{}, fmt.Errorf("workdir: write gc meta: %w", err)
	}
	return ws, nil
}

// InjectRuntimeConfig renders the provider's native config file into
// ws.Workdir. Returns an error if the provider name is empty or
// contains path traversal.
func (a *FSAdapter) InjectRuntimeConfig(ws Workspace, cfg RuntimeConfig) error {
	provider := strings.TrimSpace(cfg.Provider)
	if provider == "" {
		return errors.New("workdir: RuntimeConfig.Provider is required")
	}
	if !safePathSegment(provider) {
		return errors.New("workdir: provider name contains a path separator")
	}
	if strings.TrimSpace(ws.Workdir) == "" {
		return errors.New("workdir: workspace workdir is required")
	}

	plan, err := ResolveProviderConfigPlanWithOptions(provider, ProviderConfigPlanOptions{
		NativeHookMode:       cfg.NativeHookMode,
		NativeConfigHomeMode: cfg.NativeConfigHomeMode,
	})
	if err != nil {
		return err
	}

	content := renderRuntimeConfig(cfg)
	if err := writeNativeConfigArtifact(ws, plan.PrimaryInstructionFile, []byte(content)); err != nil {
		return err
	}
	for _, artifact := range renderProviderNativeConfigArtifacts(plan) {
		if err := writeNativeConfigArtifactWithMode(ws, artifact.Path, artifact.Content, artifact.Mode); err != nil {
			return err
		}
	}
	manifest, err := renderNativeConfigManifest(plan, cfg)
	if err != nil {
		return err
	}
	if err := writeNativeConfigArtifact(ws, plan.ManifestFile, manifest); err != nil {
		return err
	}
	return nil
}

// Archive writes the local keep-in-place archive manifest for a terminal run.
func (a *FSAdapter) Archive(ws Workspace, req ArchiveRequest) (ArchiveRecord, error) {
	if strings.TrimSpace(ws.Root) == "" {
		return ArchiveRecord{}, errors.New("workdir: workspace root is required")
	}
	if strings.TrimSpace(ws.Workdir) == "" {
		return ArchiveRecord{}, errors.New("workdir: workdir path is required")
	}
	status := strings.TrimSpace(req.ResultStatus)
	if status == "" {
		return ArchiveRecord{}, errors.New("workdir: archive result status is required")
	}
	archivedAt := req.ArchivedAt
	if archivedAt.IsZero() {
		archivedAt = time.Now().UTC()
	} else {
		archivedAt = archivedAt.UTC()
	}
	record := ArchiveRecord{
		SchemaVersion: ArchiveRecordSchemaVersion,
		WorkdirPath:   ws.Workdir,
		ArchiveURI:    localFileURI(ws.Root),
		RetentionMode: RetentionModeKeepInPlace,
		ResultStatus:  status,
		ArchivedAt:    archivedAt,
	}
	if err := os.MkdirAll(ws.Root, 0o755); err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: mkdir archive root: %w", err)
	}
	if err := writeJSONAtomic(filepath.Join(ws.Root, "archive.json"), record); err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: write archive manifest: %w", err)
	}
	return record, nil
}

// CleanupArchivedBefore deletes run roots whose archive manifest is
// keep-in-place and older than req.ArchivedBefore. Runs without
// archive.json are considered active or dirty and are never removed.
func (a *FSAdapter) CleanupArchivedBefore(ctx context.Context, req CleanupRequest) (CleanupResult, error) {
	var result CleanupResult
	root := strings.TrimSpace(a.root)
	if root == "" {
		return result, errors.New("workdir: cleanup root is required")
	}
	cutoff := req.ArchivedBefore
	if cutoff.IsZero() {
		return result, errors.New("workdir: cleanup ArchivedBefore is required")
	}
	cutoff = cutoff.UTC()
	removedAt := req.RemovedAt
	if removedAt.IsZero() {
		removedAt = time.Now().UTC()
	} else {
		removedAt = removedAt.UTC()
	}
	info, err := os.Stat(root)
	if errors.Is(err, fs.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("workdir: stat cleanup root: %w", err)
	}
	if !info.IsDir() {
		return result, fmt.Errorf("workdir: cleanup root is not a directory: %s", root)
	}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() || filepath.Base(path) != "archive.json" {
			return nil
		}
		result.ScannedArchiveRecords++
		record, err := readArchiveRecord(path)
		if err != nil {
			return err
		}
		if !cleanupEligible(record, cutoff) {
			return nil
		}
		runRoot := filepath.Dir(path)
		if runRoot == root {
			return errors.New("workdir: refusing to remove cleanup root")
		}
		if err := os.RemoveAll(runRoot); err != nil {
			return fmt.Errorf("workdir: remove archived run %s: %w", runRoot, err)
		}
		result.Removed = append(result.Removed, CleanupRecord{
			RunRoot:   runRoot,
			Archive:   record,
			RemovedAt: removedAt,
		})
		return filepath.SkipDir
	})
	if err != nil {
		return result, fmt.Errorf("workdir: cleanup archived runs: %w", err)
	}
	return result, nil
}

// ComputeNativeConfigVersion returns the sha256-hex NativeConfigVersion
// defined by docs/20-domain/workspace.md §6 for the materialized native
// config tree.
func ComputeNativeConfigVersion(ws Workspace, input NativeConfigVersionInput) (string, error) {
	if strings.TrimSpace(ws.NativeConfig) == "" {
		return "", errors.New("workdir: native-config dir is required")
	}
	if strings.TrimSpace(input.PolicyBundleVersion) == "" {
		return "", errors.New("workdir: policy bundle version is required")
	}
	if strings.TrimSpace(input.ProviderKind) == "" {
		return "", errors.New("workdir: provider kind is required")
	}
	if strings.TrimSpace(input.ProtocolKind) == "" {
		return "", errors.New("workdir: protocol kind is required")
	}
	injected, err := injectedFileHashes(ws.NativeConfig)
	if err != nil {
		return "", err
	}
	if len(injected) == 0 {
		return "", errors.New("workdir: native-config has no injected files")
	}
	doc := nativeConfigVersionDoc{
		PolicyBundleVersion: input.PolicyBundleVersion,
		NativeConfigPlan: nativeConfigPlan{
			ProviderKind:  input.ProviderKind,
			ProtocolKind:  input.ProtocolKind,
			InjectedFiles: injected,
		},
		SchemaVersion: NativeConfigVersionSchemaVersion,
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:]), nil
}

// ProviderConfigFilename returns the native config filename for a provider.
// Unknown providers fall back to AGENTS.md through the generated
// native-config plan catalog.
func ProviderConfigFilename(provider string) string {
	return ProviderConfigPlan(provider).PrimaryInstructionFile
}

type nativeConfigArtifact struct {
	Path    string
	Content []byte
	Mode    fs.FileMode
}

func renderProviderNativeConfigArtifacts(plan ProviderNativeConfigPlan) []nativeConfigArtifact {
	var artifacts []nativeConfigArtifact
	for _, path := range plan.ProviderSettingsFiles {
		switch path {
		case ".claude/settings.json":
			artifacts = append(artifacts, nativeConfigArtifact{
				Path:    path,
				Content: []byte(claudeSettingsJSON()),
				Mode:    0o644,
			})
		case ".codex/config.toml":
			artifacts = append(artifacts, nativeConfigArtifact{
				Path:    path,
				Content: []byte(codexConfigTOML()),
				Mode:    0o644,
			})
		}
	}
	for _, path := range plan.HookFiles {
		switch path {
		case ".riido/hooks/claude-audit-hook.sh":
			artifacts = append(artifacts, nativeConfigArtifact{
				Path:    path,
				Content: []byte(claudeAuditHookScript()),
				Mode:    0o755,
			})
		}
	}
	return artifacts
}

type ProviderConfigPlanOptions struct {
	NativeHookMode       string
	NativeConfigHomeMode string
}

func ResolveProviderConfigPlan(provider string, nativeHookMode string) (ProviderNativeConfigPlan, error) {
	return ResolveProviderConfigPlanWithOptions(provider, ProviderConfigPlanOptions{
		NativeHookMode: nativeHookMode,
	})
}

func ResolveProviderConfigPlanWithOptions(provider string, opts ProviderConfigPlanOptions) (ProviderNativeConfigPlan, error) {
	plan := ProviderConfigPlan(provider)
	plan, err := applyNativeHookModeDecision(plan, opts.NativeHookMode)
	if err != nil {
		return ProviderNativeConfigPlan{}, err
	}
	return applyNativeConfigHomeModeDecision(plan, opts.NativeConfigHomeMode)
}

func applyNativeHookModeDecision(plan ProviderNativeConfigPlan, mode string) (ProviderNativeConfigPlan, error) {
	mode = strings.TrimSpace(mode)
	if mode == "" || mode == plan.HookMode {
		return plan, nil
	}
	switch mode {
	case NativeConfigHookModeInstructionOnly:
		plan.HookMode = NativeConfigHookModeInstructionOnly
		plan.HookFiles = nil
		plan.ProviderSettingsFiles = removeHookSettingsFiles(plan.ProviderSettingsFiles)
		return plan, nil
	default:
		return ProviderNativeConfigPlan{}, fmt.Errorf("workdir: native hook mode %q is not supported for provider %q", mode, plan.ProviderKind)
	}
}

func applyNativeConfigHomeModeDecision(plan ProviderNativeConfigPlan, mode string) (ProviderNativeConfigPlan, error) {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return plan, nil
	}
	switch mode {
	case NativeConfigHomeModeDisabled:
		configHomeDir := plan.ConfigHomeDir
		plan.ConfigHomeDir = ""
		plan.ProviderSettingsFiles = removeConfigHomeSettingsFiles(plan.ProviderSettingsFiles, configHomeDir)
		return plan, nil
	default:
		return ProviderNativeConfigPlan{}, fmt.Errorf("workdir: native config home mode %q is not supported for provider %q", mode, plan.ProviderKind)
	}
}

func removeHookSettingsFiles(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == ".claude/settings.json" {
			continue
		}
		out = append(out, path)
	}
	return out
}

func removeConfigHomeSettingsFiles(paths []string, configHomeDir string) []string {
	configHomeDir = filepath.ToSlash(strings.TrimSpace(configHomeDir))
	if configHomeDir == "" {
		return append([]string(nil), paths...)
	}
	prefix := strings.TrimSuffix(configHomeDir, "/") + "/"
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(filepath.ToSlash(strings.TrimSpace(path)), prefix) {
			continue
		}
		out = append(out, path)
	}
	return out
}

func renderNativeConfigManifest(plan ProviderNativeConfigPlan, cfg RuntimeConfig) ([]byte, error) {
	workflow := strings.TrimSpace(cfg.Workflow)
	if workflow == "" {
		workflow = "default"
	}
	manifest := NativeConfigManifest{
		SchemaVersion:              NativeConfigManifestSchemaVersion,
		ProviderKind:               plan.ProviderKind,
		ProtocolKind:               strings.TrimSpace(cfg.ProtocolKind),
		PrimaryInstructionFile:     plan.PrimaryInstructionFile,
		ManifestFile:               plan.ManifestFile,
		HookMode:                   plan.HookMode,
		ConfigHomeDir:              plan.ConfigHomeDir,
		ProviderSettingsFiles:      sortedUniquePaths(plan.ProviderSettingsFiles),
		HookFiles:                  sortedUniquePaths(plan.HookFiles),
		TelemetryContractPlacement: strings.TrimSpace(cfg.TelemetryContractPlacement),
		Workflow:                   workflow,
		GeneratedFiles:             plan.GeneratedFiles(),
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("workdir: marshal native config manifest: %w", err)
	}
	return append(data, '\n'), nil
}

func writeNativeConfigArtifact(ws Workspace, rel string, content []byte) error {
	return writeNativeConfigArtifactWithMode(ws, rel, content, 0o644)
}

func writeNativeConfigArtifactWithMode(ws Workspace, rel string, content []byte, mode fs.FileMode) error {
	if err := writeFileUnder(ws.Workdir, rel, content, mode); err != nil {
		return err
	}
	if ws.NativeConfig != "" {
		if err := writeFileUnder(ws.NativeConfig, rel, content, mode); err != nil {
			return fmt.Errorf("workdir: write native-config copy %s: %w", rel, err)
		}
	}
	return nil
}

func writeFileUnder(root, rel string, content []byte, mode fs.FileMode) error {
	path, err := safeJoin(root, rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("workdir: mkdir %s: %w", filepath.Dir(path), err)
	}
	if mode == 0 {
		mode = 0o644
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		return fmt.Errorf("workdir: write %s: %w", path, err)
	}
	if mode&0o111 != 0 {
		if err := os.Chmod(path, mode); err != nil {
			return fmt.Errorf("workdir: chmod %s: %w", path, err)
		}
	}
	return nil
}

func safeJoin(root, rel string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("workdir: root is required")
	}
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", errors.New("workdir: relative path is required")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("workdir: relative path is absolute: %s", rel)
	}
	clean := filepath.Clean(rel)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("workdir: relative path escapes root: %s", rel)
	}
	for _, part := range strings.Split(filepath.ToSlash(clean), "/") {
		if part == ".." {
			return "", fmt.Errorf("workdir: relative path escapes root: %s", rel)
		}
	}
	return filepath.Join(root, clean), nil
}

func claudeSettingsJSON() string {
	return `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PROJECT_DIR}/.riido/hooks/claude-audit-hook.sh",
            "timeout": 30,
            "statusMessage": "Riido audit"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PROJECT_DIR}/.riido/hooks/claude-audit-hook.sh",
            "timeout": 30,
            "statusMessage": "Riido audit"
          }
        ]
      }
    ]
  }
}
`
}

func claudeAuditHookScript() string {
	return `#!/bin/sh
set -eu

project_dir="${CLAUDE_PROJECT_DIR:-$(pwd)}"
event_dir="$project_dir/.riido/hooks"
mkdir -p "$event_dir"
cat >> "$event_dir/claude-hook-events.jsonl"
printf '\n' >> "$event_dir/claude-hook-events.jsonl"
exit 0
`
}

func codexConfigTOML() string {
	return `# Managed by riido-daemon.
# Reserved for future Codex native config materialization.
# Current Codex runs use adapter-owned full-access sandbox selection instead of task-scoped CODEX_HOME.
`
}

func sortedUniquePaths(paths []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

// safePathSegment returns true if s is a non-empty string that does not
// contain path separators or upward traversal sequences. We do NOT
// trust caller-supplied workspace/task/provider names blindly; this
// guards against escapes from the per-task tree into the shared root.
func safePathSegment(s string) bool {
	if s == "" {
		return false
	}
	if strings.ContainsRune(s, os.PathSeparator) {
		return false
	}
	if strings.Contains(s, "..") {
		return false
	}
	return true
}

func localFileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: path}).String()
}

func readArchiveRecord(path string) (ArchiveRecord, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: read archive manifest: %w", err)
	}
	var record ArchiveRecord
	if err := json.Unmarshal(body, &record); err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: decode archive manifest: %w", err)
	}
	return record, nil
}

func cleanupEligible(record ArchiveRecord, cutoff time.Time) bool {
	if record.SchemaVersion != ArchiveRecordSchemaVersion {
		return false
	}
	if record.RetentionMode != RetentionModeKeepInPlace {
		return false
	}
	if record.ArchivedAt.IsZero() {
		return false
	}
	return record.ArchivedAt.UTC().Before(cutoff)
}

func injectedFileHashes(root string) ([]nativeConfigFileHash, error) {
	files := []nativeConfigFileHash{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(content)
		files = append(files, nativeConfigFileHash{
			Path:   filepath.ToSlash(rel),
			SHA256: fmt.Sprintf("%x", sum[:]),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("workdir: walk native-config: %w", err)
	}
	sortNativeConfigFiles(files)
	return files, nil
}

func sortNativeConfigFiles(files []nativeConfigFileHash) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

func writeJSONAtomic(path string, value any) error {
	return fileutil.WriteJSONAtomic(path, value)
}

func renderRuntimeConfig(cfg RuntimeConfig) string {
	var b strings.Builder
	b.WriteString("# Runtime configuration\n\n")
	if cfg.Identity != "" {
		b.WriteString("## Identity\n\n")
		b.WriteString(cfg.Identity)
		b.WriteString("\n\n")
	}
	if len(cfg.CLICatalog) > 0 {
		b.WriteString("## CLI catalog\n\n")
		for _, line := range cfg.CLICatalog {
			b.WriteString("- `")
			b.WriteString(line)
			b.WriteString("`\n")
		}
		b.WriteString("\n")
	}
	if len(cfg.HardRules) > 0 {
		b.WriteString("## Hard rules\n\n")
		for _, r := range cfg.HardRules {
			b.WriteString("- ")
			b.WriteString(r)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	workflow := cfg.Workflow
	if workflow == "" {
		workflow = "default"
	}
	fmt.Fprintf(&b, "## Workflow\n\nworkflow: %s\n", workflow)
	return b.String()
}
