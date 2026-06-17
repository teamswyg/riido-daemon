package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func (a *Actor) shutdown(ctx lifecycle.Context, runtimes []*runtimeactor.Actor, inFlight map[string]*runningTask) {
	stdCtx := ctx.Context()
	finishedAt := time.Now().UTC()
	for taskID, task := range inFlight {
		if task.handle != nil {
			_ = task.runtime.Cancel(stdCtx, task.taskID, ErrStopped.Error())
		}
		if task.cancel != nil {
			task.cancel()
			task.cancel = nil
		}
		res := a.recordTerminalResult(stdCtx, task, agentbridge.Result{
			Status:     agentbridge.ResultCancelled,
			Error:      ErrStopped.Error(),
			FinishedAt: finishedAt,
		})
		_ = a.cfg.Reporter.CompleteTask(controlplane.ContextWithTaskReport(stdCtx, task.report), task.taskID, res)
		delete(inFlight, taskID)
	}
	for _, rt := range runtimes {
		status, err := rt.Status(stdCtx)
		if err != nil || status.RuntimeID == "" {
			continue
		}
		_ = a.cfg.Source.DeregisterRuntime(stdCtx, status.RuntimeID)
	}
}

func supervisorShutdownContext(level lifecycle.ShutdownLevel) (lifecycle.Context, context.CancelFunc) {
	return lifecycle.DetachedDefaultShutdown(level)
}

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
