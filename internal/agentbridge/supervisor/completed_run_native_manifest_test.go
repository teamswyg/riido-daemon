package supervisor

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func assertCompletedNativeConfigManifest(t *testing.T, runWorkdir string) {
	t.Helper()

	manifest := readNativeConfigManifest(
		t,
		filepath.Join(runWorkdir, workdir.NativeConfigManifestPath),
	)
	if manifest.ProviderKind != "fake" ||
		manifest.ProtocolKind != "fake-unknown" ||
		manifest.PrimaryInstructionFile != "AGENTS.md" ||
		manifest.TelemetryContractPlacement != agentbridge.TelemetryPlacementPrompt ||
		manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly {
		t.Fatalf("native config manifest = %+v", manifest)
	}
}
