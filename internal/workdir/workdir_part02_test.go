package workdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectOpenClawAndCursorRemainInstructionOnly(t *testing.T) {
	for _, provider := range []string{"openclaw", "cursor"} {
		t.Run(provider, func(t *testing.T) {
			root := t.TempDir()
			a := NewFSAdapter(root)
			ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-" + provider})
			if err != nil {
				t.Fatal(err)
			}
			if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
				Provider:                   provider,
				ProtocolKind:               provider + "-protocol",
				TelemetryContractPlacement: "prompt",
				Identity:                   "id",
				Workflow:                   "default",
			}); err != nil {
				t.Fatal(err)
			}

			manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
			if manifest.ProviderKind != provider ||
				manifest.ProtocolKind != provider+"-protocol" ||
				manifest.PrimaryInstructionFile != "AGENTS.md" ||
				manifest.ManifestFile != NativeConfigManifestPath ||
				manifest.HookMode != NativeConfigHookModeInstructionOnly ||
				manifest.ConfigHomeDir != "" ||
				manifest.TelemetryContractPlacement != "prompt" ||
				manifest.Workflow != "default" ||
				len(manifest.ProviderSettingsFiles) != 0 ||
				len(manifest.HookFiles) != 0 {
				t.Fatalf("manifest = %+v", manifest)
			}
			if len(manifest.GeneratedFiles) != 2 ||
				!containsString(manifest.GeneratedFiles, "AGENTS.md") ||
				!containsString(manifest.GeneratedFiles, NativeConfigManifestPath) {
				t.Fatalf("generated files = %+v", manifest.GeneratedFiles)
			}
			for _, want := range manifest.GeneratedFiles {
				if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(want))); err != nil {
					t.Fatalf("workdir generated file %s missing: %v", want, err)
				}
				if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(want))); err != nil {
					t.Fatalf("native-config generated file %s missing: %v", want, err)
				}
			}
			for _, blocked := range []string{
				".cursor/settings.json",
				".cursor/rules",
				".openclaw/settings.json",
				".openclaw/config.json",
			} {
				if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
					t.Fatalf("provider-native artifact %s should be absent from workdir, stat err=%v", blocked, err)
				}
				if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
					t.Fatalf("provider-native artifact %s should be absent from native-config, stat err=%v", blocked, err)
				}
			}
		})
	}
}

func TestInjectClaudeWritesSettingsAndHook(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-claude"})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{Provider: "claude", ProtocolKind: "claude-stream-json"}); err != nil {
		t.Fatal(err)
	}
	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.HookMode != NativeConfigHookModeClaudeCommandHooks {
		t.Fatalf("hook mode = %q", manifest.HookMode)
	}
	for _, want := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if !containsString(manifest.GeneratedFiles, want) {
			t.Fatalf("manifest generated files missing %q: %+v", want, manifest.GeneratedFiles)
		}
	}
	if !containsString(manifest.ProviderSettingsFiles, ".claude/settings.json") {
		t.Fatalf("provider settings files = %+v", manifest.ProviderSettingsFiles)
	}
	if !containsString(manifest.HookFiles, ".riido/hooks/claude-audit-hook.sh") {
		t.Fatalf("hook files = %+v", manifest.HookFiles)
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
	hookPath := filepath.Join(ws.Workdir, ".riido", "hooks", "claude-audit-hook.sh")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook script missing: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("hook script mode = %v", info.Mode().Perm())
	}
	if _, err := os.Stat(filepath.Join(ws.NativeConfig, ".claude", "settings.json")); err != nil {
		t.Fatalf("native config claude settings copy missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.NativeConfig, ".riido", "hooks", "claude-audit-hook.sh")); err != nil {
		t.Fatalf("native config hook copy missing: %v", err)
	}
}

func TestInjectClaudeCanApplyInstructionOnlyHookPolicy(t *testing.T) {
	root := t.TempDir()
	a := NewFSAdapter(root)
	ws, err := a.Prepare(TaskID{Workspace: "ws-1", Task: "task-claude-no-hooks"})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.InjectRuntimeConfig(ws, RuntimeConfig{
		Provider:       "claude",
		ProtocolKind:   "claude-stream-json",
		NativeHookMode: NativeConfigHookModeInstructionOnly,
	}); err != nil {
		t.Fatal(err)
	}
	manifest := readNativeConfigManifest(t, filepath.Join(ws.Workdir, NativeConfigManifestPath))
	if manifest.HookMode != NativeConfigHookModeInstructionOnly {
		t.Fatalf("hook mode = %q", manifest.HookMode)
	}
	for _, blocked := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest generated files must not include blocked hook artifact %q: %+v", blocked, manifest.GeneratedFiles)
		}
		if _, err := os.Stat(filepath.Join(ws.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("workdir hook artifact %s should be absent, stat err=%v", blocked, err)
		}
		if _, err := os.Stat(filepath.Join(ws.NativeConfig, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("native-config hook artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
	if _, err := os.Stat(filepath.Join(ws.Workdir, "CLAUDE.md")); err != nil {
		t.Fatalf("primary instruction file should remain: %v", err)
	}
}
