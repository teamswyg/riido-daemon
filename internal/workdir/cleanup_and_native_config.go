package workdir

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
		if path == ".riido/hooks/claude-audit-hook.sh" {
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

func ResolveProviderConfigPlan(provider, nativeHookMode string) (ProviderNativeConfigPlan, error) {
	return ResolveProviderConfigPlanWithOptions(provider, ProviderConfigPlanOptions{
		NativeHookMode: nativeHookMode,
	})
}
