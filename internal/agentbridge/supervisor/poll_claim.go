package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
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
	for idx, rt := range runtimes {
		if claimCtx.Err() != nil {
			return claimed
		}
		status, err := rt.Status(claimCtx)
		if err != nil {
			continue
		}
		// Buffered provider events may have cancelled the previous claim before
		// this wave began. Do not block their delivery on another idle wait.
		longPoll := idx == len(runtimes)-1 && len(inFlight) == 0 && len(a.mailbox) == 0
		runtimeClaimCtx := controlplane.ContextWithClaimLongPoll(claimCtx, longPoll)
		owned := 0
		for _, task := range inFlight {
			if task.runtimeID == status.RuntimeID {
				owned++
			}
		}
		runtimeClaimCtx = controlplane.ContextWithClaimCapacity(runtimeClaimCtx, status.MaxConcurrent <= 0 || max(owned, status.RunningSessions) < status.MaxConcurrent)
		if a.claimOne(ctx, runtimeClaimCtx, rt, status, inFlight) {
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
