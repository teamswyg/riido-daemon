package supervisor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorAppliesPolicyBundleToCodexConfigHome(t *testing.T) {
	run := runPolicyBundleScenario(t, policyBundleScenarioConfig{
		provider:                 "codex",
		taskID:                   "t-codex-home",
		bundleVersion:            "policy-bundle.no-codex-home.v1",
		allowExperimentalRuntime: true,
	})

	if hasEnvPrefix(run.running.Command().Env, "TEST_NATIVE_CONFIG_HOME=") {
		t.Fatalf("native config home metadata should be omitted when policy denies it: %+v", run.running.Command())
	}
	manifest := readNativeConfigManifest(t, filepath.Join(run.result.Workdir, workdir.NativeConfigManifestPath))
	if manifest.ProviderKind != "codex" || manifest.ConfigHomeDir != "" {
		t.Fatalf("native config manifest = %+v", manifest)
	}
	if containsString(manifest.GeneratedFiles, ".codex/config.toml") {
		t.Fatalf("manifest generated files must not include Codex config home artifact: %+v", manifest.GeneratedFiles)
	}
	if _, err := os.Stat(filepath.Join(run.result.Workdir, ".codex", "config.toml")); !os.IsNotExist(err) {
		t.Fatalf("Codex config home artifact should be absent, stat err=%v", err)
	}
}
