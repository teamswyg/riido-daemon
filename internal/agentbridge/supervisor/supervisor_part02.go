package supervisor

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) Start(ctx context.Context) error {
	runtimes := configuredRuntimes(a.cfg)
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil {
			return fmt.Errorf("supervisor: runtime status: %w", err)
		}
		if err := a.register(ctx, status); err != nil {
			return err
		}
	}
	go a.run(ctx, runtimes)
	return nil
}

func configuredRuntimes(cfg Config) []*runtimeactor.Actor {
	if len(cfg.Runtimes) > 0 {
		out := make([]*runtimeactor.Actor, 0, len(cfg.Runtimes))
		for _, rt := range cfg.Runtimes {
			if rt != nil {
				out = append(out, rt)
			}
		}
		return out
	}
	if cfg.Runtime != nil {
		return []*runtimeactor.Actor{cfg.Runtime}
	}
	return nil
}

func (a *Actor) register(ctx context.Context, status runtimeactor.Status) error {
	caps := map[string]bool{}
	attrs := map[string]string{}
	for _, c := range status.Capabilities {
		prefix := "provider." + c.Provider + "."
		caps[prefix+"available"] = c.Available
		caps[prefix+"requires_experimental_opt_in"] = c.RequiresExperimentalOptIn
		caps[prefix+"supports_streaming"] = c.SupportsStreaming
		caps[prefix+"supports_resume"] = c.SupportsResume
		caps[prefix+"supports_system"] = c.SupportsSystem
		caps[prefix+"supports_max_turns"] = c.SupportsMaxTurns
		caps[prefix+"supports_mcp"] = c.SupportsMCP
		caps[prefix+"supports_tool_hooks"] = c.SupportsToolHooks
		caps[prefix+"supports_usage"] = c.SupportsUsage
		caps[prefix+"supports_file_events"] = c.SupportsFileEvents
		caps[prefix+"supports_worktree"] = c.SupportsWorktree
		attrs[prefix+"compatibility_status"] = c.CompatibilityStatus
		attrs[prefix+"capability_fingerprint"] = c.CapabilityFingerprint
		attrs[prefix+"protocol_kind"] = c.ProtocolKind
		attrs[prefix+"protocol_version"] = c.ProtocolVersion
		attrs[prefix+"adapter_id"] = c.AdapterID
		attrs[prefix+"adapter_version"] = c.AdapterVersion
		attrs[prefix+"provider_version"] = c.Version
	}
	reg := controlplane.RuntimeRegistration{
		DaemonID:             a.cfg.DaemonID,
		RuntimeID:            status.RuntimeID,
		Provider:             statusProvider(status),
		Capabilities:         caps,
		CapabilityAttributes: attrs,
		DeviceName:           status.DeviceName,
		Models:               runtimeModels(status.Models),
		StartedAt:            status.StartedAt,
		UptimeSeconds:        status.UptimeSeconds,
		SlotLimit:            status.MaxConcurrent,
		SlotsInUse:           status.RunningSessions,
		RunningTaskIDs:       runtimeTaskIDs(status.RunningTasks),
	}
	return a.cfg.Source.RegisterRuntime(ctx, reg)
}

func runtimeModels(in []runtimeactor.RuntimeModel) []controlplane.RuntimeModel {
	out := make([]controlplane.RuntimeModel, 0, len(in))
	for _, model := range in {
		out = append(out, controlplane.RuntimeModel{
			ModelID:   model.ModelID,
			Label:     model.Label,
			IsDefault: model.IsDefault,
		})
	}
	return out
}

func statusProvider(status runtimeactor.Status) string {
	if len(status.Capabilities) == 1 && status.Capabilities[0].Provider != "" {
		return status.Capabilities[0].Provider
	}
	return "multi"
}
