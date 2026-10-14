package workdir

import (
	"path/filepath"
	"testing"
)

func TestInjectCodexWritesManifest(t *testing.T) {
	ws := injectedWorkspace(t, RuntimeConfig{
		Provider:                   "codex",
		ProtocolKind:               "codex-app-server",
		TelemetryContractPlacement: "prompt",
		Identity:                   "id",
		Workflow:                   "quick-create",
	})
	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.SchemaVersion != NativeConfigManifestSchemaVersion ||
		manifest.ProviderKind != "codex" ||
		manifest.ProtocolKind != "codex-app-server" ||
		manifest.PrimaryInstructionFile != "AGENTS.md" ||
		manifest.ManifestFile != NativeConfigManifestPath ||
		manifest.HookMode != NativeConfigHookModeInstructionOnly ||
		manifest.ConfigHomeDir != "" ||
		manifest.TelemetryContractPlacement != "prompt" ||
		manifest.Workflow != "quick-create" {
		t.Fatalf("manifest = %+v", manifest)
	}
	if len(manifest.ProviderSettingsFiles) != 0 {
		t.Fatalf("manifest provider settings files = %+v", manifest.ProviderSettingsFiles)
	}
	for _, want := range []string{"AGENTS.md", NativeConfigManifestPath} {
		if !containsString(manifest.GeneratedFiles, want) {
			t.Fatalf("manifest generated files missing %q: %+v", want, manifest.GeneratedFiles)
		}
	}
}
