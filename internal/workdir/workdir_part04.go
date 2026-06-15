package workdir

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

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
	if slices.Contains(strings.Split(filepath.ToSlash(clean), "/"), "..") {
		return "", fmt.Errorf("workdir: relative path escapes root: %s", rel)
	}
	return filepath.Join(root, clean), nil
}
