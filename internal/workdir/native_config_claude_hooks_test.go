package workdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectClaudeWritesSettingsAndHook(t *testing.T) {
	ws := injectedWorkspace(t, RuntimeConfig{Provider: "claude", ProtocolKind: "claude-stream-json"})
	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.HookMode != NativeConfigHookModeClaudeCommandHooks {
		t.Fatalf("hook mode = %q", manifest.HookMode)
	}
	for _, want := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if !containsString(manifest.GeneratedFiles, want) {
			t.Fatalf("manifest generated files missing %q: %+v", want, manifest.GeneratedFiles)
		}
	}
	assertClaudeHookFiles(t, ws, manifest)
}

func assertClaudeHookFiles(t *testing.T, ws Workspace, manifest NativeConfigManifest) {
	t.Helper()
	if !containsString(manifest.ProviderSettingsFiles, ".claude/settings.json") ||
		!containsString(manifest.HookFiles, ".riido/hooks/claude-audit-hook.sh") {
		t.Fatalf("manifest hook metadata = %+v", manifest)
	}
	settings, err := os.ReadFile(filepath.Join(ws.Workdir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read claude settings: %v", err)
	}
	if !strings.Contains(string(settings), `"PreToolUse"`) ||
		!strings.Contains(string(settings), `"PostToolUse"`) ||
		!strings.Contains(string(settings), "${CLAUDE_PROJECT_DIR}/.riido/hooks/claude-audit-hook.sh") {
		t.Fatalf("claude settings missing hook config:\n%s", settings)
	}
}
