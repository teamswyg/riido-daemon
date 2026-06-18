package workdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInjectOpenClawAndCursorRemainInstructionOnly(t *testing.T) {
	for _, provider := range []string{"openclaw", "cursor"} {
		t.Run(provider, func(t *testing.T) {
			ws := injectedWorkspace(t, RuntimeConfig{
				Provider:                   provider,
				ProtocolKind:               provider + "-protocol",
				TelemetryContractPlacement: "prompt",
				Identity:                   "id",
				Workflow:                   "default",
			})
			manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
			assertInstructionOnlyManifest(t, manifest, provider)
			assertGeneratedFilesExist(t, ws, manifest.GeneratedFiles)
			assertProviderNativeArtifactsAbsent(t, ws)
		})
	}
}

func assertProviderNativeArtifactsAbsent(t *testing.T, ws Workspace) {
	t.Helper()
	for _, blocked := range []string{".cursor/settings.json", ".cursor/rules", ".openclaw/settings.json", ".openclaw/config.json"} {
		if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("workdir provider artifact %s should be absent, stat err=%v", blocked, err)
		}
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("native-config provider artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
}
