package supervisor

import (
	"context"
	"log"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) refreshRuntimeRegistrations(
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
	}
}

func (a *Actor) reportLiveness(ctx context.Context, runtimes []*runtimeactor.Actor) {
	for _, rt := range runtimes {
		requestCtx, cancel := context.WithTimeout(ctx, a.cfg.HeartbeatEvery)
		status, err := rt.Status(requestCtx)
		if err == nil {
			_ = a.cfg.Source.Heartbeat(requestCtx, controlplane.RuntimeHeartbeat{RuntimeID: status.RuntimeID, UptimeSeconds: status.UptimeSeconds, DeviceName: status.DeviceName, SlotLimit: status.MaxConcurrent, SlotsInUse: status.RunningSessions, RunningTaskIDs: a.heartbeatTaskIDs(status.RuntimeID, runtimeTaskIDs(status.RunningTasks))})
		}
		cancel()
	}
}

func (a *Actor) heartbeatLoop(ctx context.Context, runtimes []*runtimeactor.Actor) {
	ticker := time.NewTicker(a.cfg.HeartbeatEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.reportLiveness(ctx, runtimes)
		}
	}
}
