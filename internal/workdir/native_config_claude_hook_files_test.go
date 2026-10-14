package workdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInjectClaudeWritesExecutableHookCopies(t *testing.T) {
	ws := injectedWorkspace(t, RuntimeConfig{Provider: "claude", ProtocolKind: "claude-stream-json"})
	hookPath := filepath.Join(ws.Workdir, ".riido", "hooks", "claude-audit-hook.sh")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook script missing: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("hook script mode = %v", info.Mode().Perm())
	}
	for _, want := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(want))); err != nil {
			t.Fatalf("native config copy missing %s: %v", want, err)
		}
	}
}
