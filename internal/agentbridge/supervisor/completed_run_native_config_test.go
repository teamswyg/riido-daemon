package supervisor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func assertCompletedNativeConfigInjected(t *testing.T, runWorkdir string) {
	t.Helper()

	assertCompletedNativeConfigFile(t, runWorkdir)
	assertCompletedNativeConfigCopy(t, runWorkdir)
	assertCompletedNativeConfigManifest(t, runWorkdir)
	assertCompletedNativeConfigManifestCopy(t, runWorkdir)
}

func assertCompletedNativeConfigFile(t *testing.T, runWorkdir string) {
	t.Helper()

	nativeConfig, err := os.ReadFile(filepath.Join(runWorkdir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read runtime config: %v", err)
	}
	if !strings.Contains(string(nativeConfig), "<riido_log>") {
		t.Fatalf("runtime config missing telemetry hard rule:\n%s", nativeConfig)
	}
}

func assertCompletedNativeConfigCopy(t *testing.T, runWorkdir string) {
	t.Helper()

	path := filepath.Join(filepath.Dir(runWorkdir), "native-config", "AGENTS.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("native config copy not injected: %v", err)
	}
}

func assertCompletedNativeConfigManifestCopy(t *testing.T, runWorkdir string) {
	t.Helper()

	path := filepath.Join(filepath.Dir(runWorkdir), "native-config", filepath.FromSlash(workdir.NativeConfigManifestPath))
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("native config manifest copy not injected: %v", err)
	}
}
