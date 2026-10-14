package supervisor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type resolvedNativeWorkspaceConfig struct {
	hookMode       string
	configHomeMode string
	resolved       workdir.ProviderNativeConfigPlan
}

func (a *Actor) resolveNativeWorkspaceConfig(req *bridge.TaskRequest) (resolvedNativeWorkspaceConfig, error) {
	nativePlan := workdir.ProviderConfigPlan(string(req.Provider))
	native := resolvedNativeWorkspaceConfig{
		hookMode:       a.nativeHookMode(nativePlan),
		configHomeMode: a.nativeConfigHomeMode(nativePlan),
	}
	resolved, err := workdir.ResolveProviderConfigPlanWithOptions(string(req.Provider), workdir.ProviderConfigPlanOptions{
		NativeHookMode:       native.hookMode,
		NativeConfigHomeMode: native.configHomeMode,
	})
	if err != nil {
		return resolvedNativeWorkspaceConfig{}, err
	}
	native.resolved = resolved
	return native, nil
}

func (a *Actor) injectWorkspaceRuntimeConfig(ws workdir.Workspace, req *bridge.TaskRequest, protocolKind string, native resolvedNativeWorkspaceConfig) error {
	return a.cfg.Workdir.InjectRuntimeConfig(ws, workdir.RuntimeConfig{
		Provider:                   string(req.Provider),
		ProtocolKind:               protocolKind,
		TelemetryContractPlacement: req.Metadata[agentbridge.MetadataTelemetryContract],
		NativeHookMode:             native.hookMode,
		NativeConfigHomeMode:       native.configHomeMode,
		Identity:                   runtimeIdentity(req.Metadata),
		HardRules:                  runtimeHardRules(req.Metadata),
		Workflow:                   req.Metadata[MetadataWorkflow],
	})
}
