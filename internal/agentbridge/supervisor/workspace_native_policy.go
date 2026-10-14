package supervisor

import (
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) nativeHookMode(plan workdir.ProviderNativeConfigPlan) string {
	switch plan.HookMode {
	case workdir.NativeConfigHookModeClaudeCommandHooks:
		decision := policy.EvaluateNativeConfigHookWithBundle(a.cfg.PolicyBundle, policy.NativeConfigHookInput{
			TrustTier: a.cfg.RuntimeTrustTier,
			Surface:   policy.NativeConfigHookClaudeCommandAudit,
		})
		if decision.Allowed {
			return plan.HookMode
		}
		return workdir.NativeConfigHookModeInstructionOnly
	default:
		return plan.HookMode
	}
}

func (a *Actor) nativeConfigHomeMode(plan workdir.ProviderNativeConfigPlan) string {
	if providercatalog.IsCodex(plan.ProviderKind) && plan.ConfigHomeDir == ".codex" {
		decision := policy.EvaluateNativeConfigFileWithBundle(a.cfg.PolicyBundle, policy.NativeConfigFileInput{
			TrustTier: a.cfg.RuntimeTrustTier,
			Surface:   policy.NativeConfigFileCodexTaskScopedHome,
		})
		if decision.Allowed {
			return ""
		}
		return workdir.NativeConfigHomeModeDisabled
	}
	return ""
}
