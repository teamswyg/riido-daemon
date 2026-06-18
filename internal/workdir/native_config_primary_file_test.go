package workdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectClaudeWritesCLAUDEmd(t *testing.T) {
	ws := injectedWorkspace(t, RuntimeConfig{
		Provider:   "claude",
		Identity:   "Agent: tester (id: t-1)",
		CLICatalog: []string{"riido task list", "riido api status"},
		HardRules:  []string{"Use --output json always."},
		Workflow:   "default",
	})
	content := readFile(t, filepath.Join(ws.Workdir, "CLAUDE.md"))
	for _, want := range []string{"Agent: tester (id: t-1)", "riido task list", "Use --output json always.", "workflow: default"} {
		if !strings.Contains(content, want) {
			t.Fatalf("CLAUDE.md missing %q:\n%s", want, content)
		}
	}
}

func TestInjectCodexWritesAGENTSmd(t *testing.T) {
	ws := injectedWorkspace(t, RuntimeConfig{
		Provider:                   "codex",
		ProtocolKind:               "codex-app-server",
		TelemetryContractPlacement: "prompt",
		Identity:                   "id",
		Workflow:                   "quick-create",
	})
	if _, err := os.Stat(filepath.Join(ws.Workdir, "AGENTS.md")); err != nil {
		t.Fatalf("AGENTS.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.NativeConfig, "AGENTS.md")); err != nil {
		t.Fatalf("native-config AGENTS.md copy missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "CLAUDE.md")); err == nil {
		t.Fatalf("codex must not create CLAUDE.md")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(bytes)
}
