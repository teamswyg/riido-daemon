package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) reportRuntimeHeartbeats(ctx context.Context, runtimes []*runtimeactor.Actor) {
	for _, rt := range runtimes {
		hb, err := rt.HeartbeatPayload(ctx)
		if err != nil {
			continue
		}
		_ = a.cfg.Source.Heartbeat(ctx, controlplane.RuntimeHeartbeat{
			RuntimeID:      hb.RuntimeID,
			UptimeSeconds:  hb.UptimeSeconds,
			DeviceName:     hb.DeviceName,
			SlotLimit:      hb.SlotLimit,
			SlotsInUse:     hb.SlotsInUse,
			RunningTaskIDs: hb.RunningTaskIDs,
		})
	}
}
