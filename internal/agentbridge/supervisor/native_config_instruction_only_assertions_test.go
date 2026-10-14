package supervisor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func assertNativeConfigHomeMetadataOmitted(t *testing.T, run nativeConfigInstructionOnlyRun) {
	t.Helper()
	if hasEnvPrefix(run.command.Env, "TEST_NATIVE_CONFIG_HOME=") {
		t.Fatalf("native config home metadata must be omitted for %s: %+v", run.provider, run.command)
	}
}

func assertNativeConfigInstructionOnlyManifest(t *testing.T, run nativeConfigInstructionOnlyRun) {
	t.Helper()
	manifest := readNativeConfigManifest(
		t,
		filepath.Join(run.result.Workdir, workdir.NativeConfigManifestPath),
	)
	if manifest.ProviderKind != string(run.provider) ||
		manifest.PrimaryInstructionFile != "AGENTS.md" ||
		manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly ||
		manifest.ConfigHomeDir != "" ||
		len(manifest.ProviderSettingsFiles) != 0 ||
		len(manifest.HookFiles) != 0 {
		t.Fatalf("native config manifest = %+v", manifest)
	}
	assertInstructionOnlyGeneratedFiles(t, manifest.GeneratedFiles)
}

func assertInstructionOnlyGeneratedFiles(t *testing.T, files []string) {
	t.Helper()
	if len(files) != 2 ||
		!containsString(files, "AGENTS.md") ||
		!containsString(files, workdir.NativeConfigManifestPath) {
		t.Fatalf("generated files = %+v", files)
	}
}

func assertProviderNativeArtifactsAbsent(t *testing.T, run nativeConfigInstructionOnlyRun) {
	t.Helper()
	for _, blocked := range providerNativeArtifactsBlockedInInstructionOnlyMode() {
		path := filepath.Join(run.result.Workdir, filepath.FromSlash(blocked))
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("provider-native artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
}

func providerNativeArtifactsBlockedInInstructionOnlyMode() []string {
	return []string{
		".cursor/settings.json",
		".cursor/rules",
		".openclaw/settings.json",
		".openclaw/config.json",
	}
}
