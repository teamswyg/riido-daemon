package supervisor

import (
	"context"
	"errors"
	"log"
	"net"
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
				logRuntimeSyncFailure("provider_health_sync_failed", status.RuntimeID, err)
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
			err = a.cfg.Source.Heartbeat(requestCtx, controlplane.RuntimeHeartbeat{RuntimeID: status.RuntimeID, UptimeSeconds: status.UptimeSeconds, DeviceName: status.DeviceName, SlotLimit: status.MaxConcurrent, SlotsInUse: status.RunningSessions, RunningTaskIDs: a.heartbeatTaskIDs(status.RuntimeID, runtimeTaskIDs(status.RunningTasks))})
		}
		if err != nil {
			logRuntimeSyncFailure("heartbeat_sync_failed", status.RuntimeID, err)
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

func logRuntimeSyncFailure(event, runtimeID string, err error) {
	kind, status := runtimeSyncFailure(err)
	log.Printf("riido_daemon event=%s runtime_id=%q failure_kind=%s http_status=%d", event, runtimeID, kind, status)
}

func runtimeSyncFailure(err error) (string, int) {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout", 0
	}
	if errors.Is(err, context.Canceled) {
		return "cancelled", 0
	}
	var status interface{ HTTPStatusCode() int }
	if errors.As(err, &status) {
		return "http", status.HTTPStatusCode()
	}
	var network net.Error
	if errors.As(err, &network) {
		if network.Timeout() {
			return "timeout", 0
		}
		return "network", 0
	}
	return "internal", 0
}
