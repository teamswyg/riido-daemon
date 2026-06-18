package workdir

import "testing"

func TestProviderConfigPlanUsesGeneratedCatalog(t *testing.T) {
	if NativeConfigPlanSchemaVersion != "riido-native-config-plan.v1" {
		t.Fatalf("native config plan schema = %q", NativeConfigPlanSchemaVersion)
	}
	assertClaudePlan(t, ProviderConfigPlan(" Claude "))
	assertGenericInstructionPlan(t, "codex", ProviderConfigPlan("codex"))
	assertGenericInstructionPlan(t, "cursor", ProviderConfigPlan("cursor"))
	assertGenericInstructionPlan(t, "openclaw", ProviderConfigPlan("openclaw"))
	unknown := ProviderConfigPlan("unknown-provider")
	if unknown.ProviderKind != "unknown-provider" ||
		unknown.PrimaryInstructionFile != "AGENTS.md" ||
		unknown.ManifestFile != NativeConfigManifestPath {
		t.Fatalf("unknown plan = %+v", unknown)
	}
}

func assertClaudePlan(t *testing.T, plan ProviderNativeConfigPlan) {
	t.Helper()
	if plan.ProviderKind != "claude" ||
		plan.PrimaryInstructionFile != "CLAUDE.md" ||
		plan.HookMode != NativeConfigHookModeClaudeCommandHooks ||
		!containsString(plan.ProviderSettingsFiles, ".claude/settings.json") ||
		!containsString(plan.HookFiles, ".riido/hooks/claude-audit-hook.sh") {
		t.Fatalf("claude plan = %+v", plan)
	}
}

func assertGenericInstructionPlan(t *testing.T, provider string, plan ProviderNativeConfigPlan) {
	t.Helper()
	if plan.ProviderKind != provider ||
		plan.PrimaryInstructionFile != "AGENTS.md" ||
		plan.HookMode != NativeConfigHookModeInstructionOnly ||
		plan.ConfigHomeDir != "" ||
		len(plan.ProviderSettingsFiles) != 0 ||
		len(plan.HookFiles) != 0 {
		t.Fatalf("%s plan = %+v", provider, plan)
	}
}
