package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) claimAvailable(
	ctx context.Context,
	runtimes []*runtimeactor.Actor,
	inFlight map[string]*runningTask,
) bool {
	claimCtx, release := a.beginClaim(ctx)
	defer release()

	claimed := false
	for _, rt := range runtimes {
		if claimCtx.Err() != nil {
			return claimed
		}
		status, err := rt.Status(claimCtx)
		if err != nil {
			continue
		}
		if a.claimOne(ctx, claimCtx, rt, status, inFlight) {
			claimed = true
		}
	}
	return claimed
}

func (a *Actor) beginClaim(ctx context.Context) (context.Context, func()) {
	claimCtx, cancel := context.WithCancel(ctx)
	a.claimMu.Lock()
	a.claimCancel = cancel
	a.claimMu.Unlock()

	return claimCtx, func() {
		cancel()
		a.claimMu.Lock()
		a.claimCancel = nil
		a.claimMu.Unlock()
	}
}

func (a *Actor) cancelCurrentClaim() {
	a.claimMu.Lock()
	cancel := a.claimCancel
	a.claimMu.Unlock()
	if cancel != nil {
		cancel()
	}
}
