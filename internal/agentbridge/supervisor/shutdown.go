package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func (a *Actor) shutdown(ctx lifecycle.Context, runtimes []*runtimeactor.Actor, inFlight map[string]*runningTask) {
	stdCtx := ctx.Context()
	finishedAt := time.Now().UTC()
	a.cancelInFlightTasks(stdCtx, inFlight, finishedAt)
	a.deregisterRuntimes(stdCtx, runtimes)
}

func supervisorShutdownContext(level lifecycle.ShutdownLevel) (lifecycle.Context, context.CancelFunc) {
	return lifecycle.DetachedDefaultShutdown(level)
}
