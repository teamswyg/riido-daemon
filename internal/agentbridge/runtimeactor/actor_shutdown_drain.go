package runtimeactor

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func (a *Actor) drainAndShutdown(level lifecycle.ShutdownLevel, inFlight map[string]*runningTask, completeCh <-chan string) {
	shutdownCtx, cancel := lifecycle.DetachedDefaultShutdown(level)
	defer cancel()
	for _, t := range inFlight {
		t.handle.session.CancelWithContext(shutdownCtx.Context(), ErrActorStopped)
	}
	if level.IsForced() {
		a.stopErrCh <- nil
		return
	}
	a.waitForShutdownDrain(inFlight, completeCh)
}

func (a *Actor) waitForShutdownDrain(inFlight map[string]*runningTask, completeCh <-chan string) {
	deadline := time.After(5 * time.Second)
	for len(inFlight) > 0 {
		select {
		case id := <-completeCh:
			delete(inFlight, id)
		case next := <-a.stopReqCh:
			if lifecycle.NormalizeShutdownLevel(next).IsForced() {
				a.forceCancelInFlight(inFlight)
				a.stopErrCh <- nil
				return
			}
		case <-deadline:
			a.stopErrCh <- fmt.Errorf("runtimeactor: %d session(s) did not terminate", len(inFlight))
			return
		}
	}
	a.stopErrCh <- nil
}

func (a *Actor) forceCancelInFlight(inFlight map[string]*runningTask) {
	forcedCtx, forcedCancel := lifecycle.DetachedDefaultShutdown(lifecycle.ShutdownForced)
	defer forcedCancel()
	for _, t := range inFlight {
		t.handle.session.CancelWithContext(forcedCtx.Context(), ErrActorStopped)
	}
}
