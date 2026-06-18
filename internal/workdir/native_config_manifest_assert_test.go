package workdir

import (
	"os"
	"path/filepath"
	"testing"
)

func assertInstructionOnlyManifest(t *testing.T, manifest NativeConfigManifest, provider string) {
	t.Helper()
	if manifest.ProviderKind != provider ||
		manifest.ProtocolKind != provider+"-protocol" ||
		manifest.PrimaryInstructionFile != "AGENTS.md" ||
		manifest.ManifestFile != NativeConfigManifestPath ||
		manifest.HookMode != NativeConfigHookModeInstructionOnly ||
		manifest.ConfigHomeDir != "" ||
		manifest.TelemetryContractPlacement != "prompt" ||
		manifest.Workflow != "default" ||
		len(manifest.ProviderSettingsFiles) != 0 ||
		len(manifest.HookFiles) != 0 {
		t.Fatalf("manifest = %+v", manifest)
	}
}

func assertGeneratedFilesExist(t *testing.T, ws Workspace, files []string) {
	t.Helper()
	for _, want := range files {
		if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(want))); err != nil {
			t.Fatalf("workdir generated file %s missing: %v", want, err)
		}
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(want))); err != nil {
			t.Fatalf("native-config generated file %s missing: %v", want, err)
		}
	}
}
