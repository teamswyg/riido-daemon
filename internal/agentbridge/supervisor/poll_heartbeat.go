package supervisor

import (
	"context"
	"log"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) reportRuntimeHeartbeats(
	ctx context.Context,
	runtimes []*runtimeactor.Actor,
	inFlight map[string]*runningTask,
) {
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil {
			continue
		}
		a.blockPreparingRuntimeDrift(ctx, inFlight, status)
		state := runtimeRegistrationState(status)
		if state != a.registeredRuntimeStates[status.RuntimeID] {
			if err := a.register(ctx, status); err != nil {
				log.Printf("riido_daemon event=provider_health_sync_failed runtime_id=%q", status.RuntimeID)
			} else {
				a.registeredRuntimeStates[status.RuntimeID] = state
			}
		}
		_ = a.cfg.Source.Heartbeat(ctx, controlplane.RuntimeHeartbeat{
			RuntimeID:      status.RuntimeID,
			UptimeSeconds:  status.UptimeSeconds,
			DeviceName:     status.DeviceName,
			SlotLimit:      status.MaxConcurrent,
			SlotsInUse:     status.RunningSessions,
			RunningTaskIDs: runtimeTaskIDs(status.RunningTasks),
		})
	}
}
