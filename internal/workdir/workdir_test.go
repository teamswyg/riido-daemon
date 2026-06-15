package workdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectClaudeWritesCLAUDEmd(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-A"})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider: "claude",
		Identity: "Agent: tester (id: t-1)",
		CLICatalog: []string{
			"riido task list",
			"riido api status",
		},
		HardRules: []string{
			"Use --output json always.",
		},
		Workflow: "default",
	}); err != nil {
		t.Fatalf("InjectRuntimeConfig: %v", err)
	}

	path := filepath.Join(ws.Workdir, "CLAUDE.md")
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(bytes)
	for _, want := range []string{
		"Agent: tester (id: t-1)",
		"riido task list",
		"Use --output json always.",
		"workflow: default",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("CLAUDE.md missing %q:\n%s", want, content)
		}
	}
}

func TestInjectCodexWritesAGENTSmd(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, _ := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-B"})
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider:                   "codex",
		ProtocolKind:               "codex-app-server",
		TelemetryContractPlacement: "prompt",
		Identity:                   "id",
		Workflow:                   "quick-create",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "AGENTS.md")); err != nil {
		t.Fatalf("AGENTS.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.NativeConfig, "AGENTS.md")); err != nil {
		t.Fatalf("native-config AGENTS.md copy missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "CLAUDE.md")); err == nil {
		t.Fatalf("codex must not create CLAUDE.md")
	}

	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.SchemaVersion != NativeConfigManifestSchemaVersion {
		t.Fatalf("manifest schema = %q", manifest.SchemaVersion)
	}
	if manifest.ProviderKind != "codex" ||
		manifest.ProtocolKind != "codex-app-server" ||
		manifest.PrimaryInstructionFile != "AGENTS.md" ||
		manifest.ManifestFile != NativeConfigManifestPath ||
		manifest.HookMode != NativeConfigHookModeInstructionOnly ||
		manifest.ConfigHomeDir != "" ||
		manifest.TelemetryContractPlacement != "prompt" ||
		manifest.Workflow != "quick-create" {
		t.Fatalf("manifest = %+v", manifest)
	}
	if len(manifest.ProviderSettingsFiles) != 0 {
		t.Fatalf("manifest provider settings files = %+v", manifest.ProviderSettingsFiles)
	}
	for _, want := range []string{"AGENTS.md", NativeConfigManifestPath} {
		if !containsString(manifest.GeneratedFiles, want) {
			t.Fatalf("manifest generated files missing %q: %+v", want, manifest.GeneratedFiles)
		}
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(want))); err != nil {
			t.Fatalf("native-config copy missing %s: %v", want, err)
		}
	}
}

func TestInjectCodexCanApplyConfigHomePolicy(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, _ := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-codex-no-home"})
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider:             "codex",
		ProtocolKind:         "codex-app-server",
		NativeConfigHomeMode: NativeConfigHomeModeDisabled,
	}); err != nil {
		t.Fatal(err)
	}

	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.ProviderKind != "codex" || manifest.ConfigHomeDir != "" {
		t.Fatalf("manifest = %+v", manifest)
	}
	for _, blocked := range []string{".codex/config.toml"} {
		if containsString(manifest.ProviderSettingsFiles, blocked) || containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest must not include blocked config home artifact %q: %+v", blocked, manifest)
		}
		if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("workdir config home artifact %s should be absent, stat err=%v", blocked, err)
		}
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("native-config config home artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "AGENTS.md")); err != nil {
		t.Fatalf("primary instruction file should remain: %v", err)
	}
}
