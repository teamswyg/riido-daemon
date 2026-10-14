package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// Submit posts a TaskRequest to the actor. The stoppedCh wait guard prevents
// callers from blocking after a buffered mailbox send races with actor stop.
func (a *Actor) Submit(ctx context.Context, req bridge.TaskRequest) (*SessionHandle, error) {
	reply := make(chan submitReply, 1)
	select {
	case a.mailbox <- envelope{submit: &submitMsg{ctx: ctx, req: req, reply: reply}}:
	case <-a.stoppedCh:
		return nil, ErrActorStopped
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.handle, res.err
	case <-a.stoppedCh:
		return nil, ErrActorStopped
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
