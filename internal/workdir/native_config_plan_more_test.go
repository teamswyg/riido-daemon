package workdir

import (
	"strings"
	"testing"
)

func TestResolveProviderConfigPlanAppliesHookModeWrapper(t *testing.T) {
	plan, err := ResolveProviderConfigPlan(" Claude ", NativeConfigHookModeInstructionOnly)
	if err != nil {
		t.Fatalf("resolve plan: %v", err)
	}
	if plan.ProviderKind != "claude" || plan.HookMode != NativeConfigHookModeInstructionOnly {
		t.Fatalf("plan = %+v", plan)
	}
	for _, blocked := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if containsString(plan.ProviderSettingsFiles, blocked) || containsString(plan.HookFiles, blocked) {
			t.Fatalf("instruction-only plan kept hook artifact %q: %+v", blocked, plan)
		}
	}
}

func TestResolveProviderConfigPlanRejectsUnsupportedHookMode(t *testing.T) {
	_, err := ResolveProviderConfigPlan("claude", "native-all")
	if err == nil || !strings.Contains(err.Error(), `native hook mode "native-all"`) {
		t.Fatalf("expected unsupported hook mode error, got %v", err)
	}
}

func TestApplyNativeConfigHomeModeDecisionRemovesOnlyConfigHomeFiles(t *testing.T) {
	plan := ProviderNativeConfigPlan{
		ProviderKind:  "codex",
		ConfigHomeDir: ".codex",
		ProviderSettingsFiles: []string{
			".codex/config.toml",
			" .codex/auth.json ",
			"AGENTS.md",
		},
	}
	got, err := applyNativeConfigHomeModeDecision(plan, NativeConfigHomeModeDisabled)
	if err != nil {
		t.Fatalf("apply config home mode: %v", err)
	}
	if got.ConfigHomeDir != "" {
		t.Fatalf("config home should be cleared: %+v", got)
	}
	if len(got.ProviderSettingsFiles) != 1 || got.ProviderSettingsFiles[0] != "AGENTS.md" {
		t.Fatalf("provider settings files = %+v", got.ProviderSettingsFiles)
	}
}

func TestCodexConfigTOMLDocumentsAdapterOwnedSandbox(t *testing.T) {
	body := codexConfigTOML()
	for _, want := range []string{"Managed by riido-daemon", "adapter-owned full-access sandbox"} {
		if !strings.Contains(body, want) {
			t.Fatalf("codex config missing %q: %q", want, body)
		}
	}
}
