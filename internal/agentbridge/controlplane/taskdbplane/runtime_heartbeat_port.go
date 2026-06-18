package taskdbplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (p *Plane) Heartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		rec, ok := p.runtimes[hb.RuntimeID]
		if !ok {
			return planeErrorf(ErrTaskDBPlaneRuntime, "heartbeat", "heartbeat for unknown runtime %q", hb.RuntimeID)
		}
		rec.LastHeartbeat = p.now().UTC()
		applyHeartbeat(&rec.RuntimeRegistration, hb)
		p.runtimes[hb.RuntimeID] = rec
		if err := p.saveRuntimeRegistry(); err != nil {
			return err
		}
		leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
		if err != nil {
			return err
		}
		leases, changed := refreshRuntimeLeases(leases, rec, hb.RunningTaskIDs, rec.LastHeartbeat, p.leaseTTL)
		if !changed {
			return nil
		}
		return saveRuntimeLeaseRegistry(p.leasePath, p.path, leases, rec.LastHeartbeat)
	})
}
