package supervisor

import "context"

type idleStopRequest struct {
	ctx   context.Context
	reply chan bool
}

// PrepareIdleUpdate checks the supervisor-owned execution and delivery state in
// the same loop that claims work. Once accepted, no new claim can race shutdown.
// The short fence expires if the caller disappears before requesting shutdown.
func (a *Actor) PrepareIdleUpdate(ctx context.Context) bool {
	reply := make(chan bool, 1)
	select {
	case a.idleStops <- idleStopRequest{ctx: ctx, reply: reply}:
	case <-ctx.Done():
		return false
	case <-a.stoppedCh:
		return false
	}
	select {
	case ready := <-reply:
		return ready
	case <-ctx.Done():
		return false
	case <-a.stoppedCh:
		return false
	}
}

func idleForUpdate(inFlight map[string]*runningTask, detached []detachedReport, mailboxPending int) bool {
	// inFlight includes preparation, verification, and unacknowledged terminal
	// reports, not just native provider processes.
	return len(inFlight) == 0 && len(detached) == 0 && mailboxPending == 0
}
