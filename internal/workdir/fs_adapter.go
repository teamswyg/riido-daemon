package workdir

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

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
		metadatakeys.WorkspaceID.String(): id.Workspace,
		metadatakeys.TaskID.String():      id.Task,
		metadatakeys.RunID.String():       runID,
		"created_at":                      time.Now().UTC().Format(time.RFC3339Nano),
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return Workspace{}, fmt.Errorf("workdir: marshal gc meta: %w", err)
	}
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
