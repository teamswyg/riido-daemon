package workdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInjectCodexCanApplyConfigHomePolicy(t *testing.T) {
	ws := injectedWorkspace(t, RuntimeConfig{
		Provider:             "codex",
		ProtocolKind:         "codex-app-server",
		NativeConfigHomeMode: NativeConfigHomeModeDisabled,
	})
	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.ProviderKind != "codex" || manifest.ConfigHomeDir != "" {
		t.Fatalf("manifest = %+v", manifest)
	}
	for _, blocked := range []string{".codex/config.toml"} {
		if containsString(manifest.ProviderSettingsFiles, blocked) || containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest must not include blocked config home artifact %q: %+v", blocked, manifest)
		}
		assertAbsent(t, ws, blocked)
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "AGENTS.md")); err != nil {
		t.Fatalf("primary instruction file should remain: %v", err)
	}
}
