package workdir

import (
	"errors"
	"strings"
)

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
	if err := writeNativeConfigArtifact(ws, plan.PrimaryInstructionFile, []byte(renderRuntimeConfig(cfg))); err != nil {
		return err
	}
	if err := writeProviderNativeConfigArtifacts(ws, plan); err != nil {
		return err
	}
	manifest, err := renderNativeConfigManifest(plan, cfg)
	if err != nil {
		return err
	}
	return writeNativeConfigArtifact(ws, plan.ManifestFile, manifest)
}

func writeProviderNativeConfigArtifacts(ws Workspace, plan ProviderNativeConfigPlan) error {
	for _, artifact := range renderProviderNativeConfigArtifacts(plan) {
		if err := writeNativeConfigArtifactWithMode(ws, artifact.Path, artifact.Content, artifact.Mode); err != nil {
			return err
		}
	}
	return nil
}
