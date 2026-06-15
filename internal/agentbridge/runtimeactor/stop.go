package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// Stop initiates graceful shutdown. Safe to call concurrently and
// repeatedly. When a lifecycle shutdown level is embedded in ctx, that
// level is honored; otherwise Stop defaults to graceful shutdown.
func (a *Actor) Stop(ctx context.Context) error {
	return a.StopLifecycle(lifecycle.StopContext(ctx))
}

// StopLifecycle initiates shutdown with an explicit lifecycle authority
// level. Forced shutdown skips the graceful drain wait; a forced request can
// also escalate an already-draining graceful stop.
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
	select {
	case <-a.stoppedCh:
		// stopErrCh holds the actor goroutine's exit error. Only one
		// shutdown path writes it, so the first reader gets the real
		// value; subsequent Stop callers reach this branch via
		// stoppedCh and return nil (their work is already done).
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
