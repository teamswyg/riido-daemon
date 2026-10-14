package workdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInjectClaudeCanApplyInstructionOnlyHookPolicy(t *testing.T) {
	ws := injectedWorkspace(t, RuntimeConfig{
		Provider:       "claude",
		ProtocolKind:   "claude-stream-json",
		NativeHookMode: NativeConfigHookModeInstructionOnly,
	})
	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.HookMode != NativeConfigHookModeInstructionOnly {
		t.Fatalf("hook mode = %q", manifest.HookMode)
	}
	for _, blocked := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest generated files must not include hook artifact %q: %+v", blocked, manifest)
		}
		assertAbsent(t, ws, blocked)
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "CLAUDE.md")); err != nil {
		t.Fatalf("primary instruction file should remain: %v", err)
	}
}

func assertAbsent(t *testing.T, ws Workspace, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Fatalf("workdir artifact %s should be absent, stat err=%v", rel, err)
	}
	if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Fatalf("native-config artifact %s should be absent, stat err=%v", rel, err)
	}
}
