package workdir

import (
	"fmt"
	"strings"
)

func ResolveProviderConfigPlanWithOptions(provider string, opts ProviderConfigPlanOptions) (ProviderNativeConfigPlan, error) {
	plan := ProviderConfigPlan(provider)
	plan, err := applyNativeHookModeDecision(plan, opts.NativeHookMode)
	if err != nil {
		return ProviderNativeConfigPlan{}, err
	}
	return applyNativeConfigHomeModeDecision(plan, opts.NativeConfigHomeMode)
}

func applyNativeHookModeDecision(plan ProviderNativeConfigPlan, mode string) (ProviderNativeConfigPlan, error) {
	mode = strings.TrimSpace(mode)
	if mode == "" || mode == plan.HookMode {
		return plan, nil
	}
	switch mode {
	case NativeConfigHookModeInstructionOnly:
		plan.HookMode = NativeConfigHookModeInstructionOnly
		plan.HookFiles = nil
		plan.ProviderSettingsFiles = removeHookSettingsFiles(plan.ProviderSettingsFiles)
		return plan, nil
	default:
		return ProviderNativeConfigPlan{}, fmt.Errorf("workdir: native hook mode %q is not supported for provider %q", mode, plan.ProviderKind)
	}
}

func applyNativeConfigHomeModeDecision(plan ProviderNativeConfigPlan, mode string) (ProviderNativeConfigPlan, error) {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return plan, nil
	}
	switch mode {
	case NativeConfigHomeModeDisabled:
		configHomeDir := plan.ConfigHomeDir
		plan.ConfigHomeDir = ""
		plan.ProviderSettingsFiles = removeConfigHomeSettingsFiles(plan.ProviderSettingsFiles, configHomeDir)
		return plan, nil
	default:
		return ProviderNativeConfigPlan{}, fmt.Errorf("workdir: native config home mode %q is not supported for provider %q", mode, plan.ProviderKind)
	}
}
