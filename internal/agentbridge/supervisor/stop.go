package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func (a *Actor) Stop(ctx context.Context) error {
	return a.StopLifecycle(lifecycle.StopContext(ctx))
}

func (a *Actor) StopLifecycle(ctx lifecycle.Context) error {
	select {
	case <-a.stoppedCh:
		return nil
	default:
	}
	select {
	case a.stopReqCh <- lifecycle.NormalizeShutdownLevel(ctx.ShutdownLevel()):
	default:
	}
	a.cancelCurrentClaim()
	select {
	case <-a.stoppedCh:
		select {
		case err := <-a.stopErrCh:
			return err
		default:
			return nil
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *Actor) drainStopLevel(level lifecycle.ShutdownLevel) lifecycle.ShutdownLevel {
	level = lifecycle.NormalizeShutdownLevel(level)
	for {
		select {
		case next := <-a.stopReqCh:
			if next.AtLeast(level) {
				level = lifecycle.NormalizeShutdownLevel(next)
			}
		default:
			return level
		}
	}
}
