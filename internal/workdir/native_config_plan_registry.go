package workdir

import "strings"

// ProviderConfigPlan returns the provider-native config files currently
// materialized by C6. Unknown providers fall back to the generic AGENTS.md
// primary instruction file.
func ProviderConfigPlan(provider string) ProviderNativeConfigPlan {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if plan, ok := providerConfigPlans[provider]; ok {
		return cloneProviderNativeConfigPlan(plan)
	}
	plan := cloneProviderNativeConfigPlan(defaultProviderConfigPlan)
	plan.ProviderKind = provider
	return plan
}

func providerNativeConfigPlan(provider, primary, manifest, hookMode, configHome string, settings, hooks, extra []string) ProviderNativeConfigPlan {
	return ProviderNativeConfigPlan{
		ProviderKind:           provider,
		PrimaryInstructionFile: primary,
		ManifestFile:           manifest,
		HookMode:               hookMode,
		ConfigHomeDir:          configHome,
		ProviderSettingsFiles:  settings,
		HookFiles:              hooks,
		ExtraFiles:             extra,
	}
}

func cloneProviderNativeConfigPlan(in ProviderNativeConfigPlan) ProviderNativeConfigPlan {
	out := in
	out.ProviderSettingsFiles = append([]string(nil), in.ProviderSettingsFiles...)
	out.HookFiles = append([]string(nil), in.HookFiles...)
	out.ExtraFiles = append([]string(nil), in.ExtraFiles...)
	return out
}
