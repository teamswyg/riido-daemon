package supervisor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorAppliesPolicyBundleToClaudeNativeHooks(t *testing.T) {
	run := runPolicyBundleScenario(t, policyBundleScenarioConfig{
		provider:      "claude",
		taskID:        "t-claude-hooks",
		bundleVersion: "policy-bundle.no-hooks.v1",
	})

	manifest := readNativeConfigManifest(t, filepath.Join(run.result.Workdir, workdir.NativeConfigManifestPath))
	if manifest.ProviderKind != "claude" || manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly {
		t.Fatalf("native config manifest = %+v", manifest)
	}
	assertPolicyBundleFilesAbsent(t, run.result.Workdir,
		".claude/settings.json",
		".riido/hooks/claude-audit-hook.sh",
	)
}

func assertPolicyBundleFilesAbsent(t *testing.T, root string, paths ...string) {
	t.Helper()
	manifest := readNativeConfigManifest(t, filepath.Join(root, workdir.NativeConfigManifestPath))
	for _, blocked := range paths {
		if containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest generated files must not include %q: %+v", blocked, manifest.GeneratedFiles)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("blocked hook artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
}
