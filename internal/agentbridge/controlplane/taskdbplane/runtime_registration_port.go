package taskdbplane

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (p *Plane) RegisterRuntime(ctx context.Context, rt controlplane.RuntimeRegistration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if rt.RuntimeID == "" {
		return planeErrorf(ErrTaskDBPlaneRuntime, "register-runtime", "empty RuntimeID")
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		p.runtimes[rt.RuntimeID] = controlplane.RegisteredRuntime{
			RuntimeRegistration: rt,
			LastHeartbeat:       p.now().UTC(),
		}
		return p.saveRuntimeRegistry()
	})
}

func (p *Plane) DeregisterRuntime(ctx context.Context, runtimeID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if runtimeID == "" {
		return planeErrorf(ErrTaskDBPlaneRuntime, "deregister-runtime", "empty RuntimeID")
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		delete(p.runtimes, runtimeID)
		return p.saveRuntimeRegistry()
	})
}
