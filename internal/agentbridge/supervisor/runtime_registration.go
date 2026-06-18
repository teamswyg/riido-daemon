package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) register(ctx context.Context, status runtimeactor.Status) error {
	reg := runtimeRegistration(a.cfg.DaemonID, status)
	return a.cfg.Source.RegisterRuntime(ctx, reg)
}

func runtimeRegistration(daemonID string, status runtimeactor.Status) controlplane.RuntimeRegistration {
	caps, attrs := runtimeCapabilityMaps(status.Capabilities)
	return controlplane.RuntimeRegistration{
		DaemonID:             daemonID,
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
