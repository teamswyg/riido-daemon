package saasplane

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (p *Plane) refreshRegisteredRuntimeSnapshot(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	now := time.Now()
	var runtimes []RuntimeSnapshotRecord
	var deviceName string
	err := p.withState(ctx, func(s *planeState) {
		if len(s.registeredRuntimes) == 0 {
			fallback, ok := runtimeSnapshotFromHeartbeat(hb)
			if !ok {
				return
			}
			s.registeredRuntimes[fallback.RuntimeID] = fallback
		}
		if !s.lastRuntimeSnapshotSync.IsZero() && now.Sub(s.lastRuntimeSnapshotSync) < runtimeSnapshotHeartbeatMinInterval {
			return
		}
		s.lastRuntimeSnapshotSync = now
		deviceName = s.registeredDeviceName
		runtimes = sortedRuntimeSnapshots(s.registeredRuntimes)
	})
	if err != nil || len(runtimes) == 0 {
		return err
	}
	return p.postRuntimeSnapshot(ctx, runtimes, deviceName)
}
