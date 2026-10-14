package workdir

import (
	"encoding/json"
	"fmt"
	"strings"
)

func renderNativeConfigManifest(plan ProviderNativeConfigPlan, cfg RuntimeConfig) ([]byte, error) {
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
		Workflow:                   nativeConfigWorkflow(cfg.Workflow),
		GeneratedFiles:             plan.GeneratedFiles(),
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("workdir: marshal native config manifest: %w", err)
	}
	return append(data, '\n'), nil
}

func nativeConfigWorkflow(workflow string) string {
	workflow = strings.TrimSpace(workflow)
	if workflow == "" {
		return "default"
	}
	return workflow
}
