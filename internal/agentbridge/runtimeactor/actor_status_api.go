package runtimeactor

import "context"

// Status returns a synchronous status snapshot.
func (a *Actor) Status(ctx context.Context) (Status, error) {
	reply := make(chan statusReply, 1)
	select {
	case a.statusCh <- statusMsg{ctx: ctx, reply: reply}:
	case <-a.stoppedCh:
		return Status{RuntimeID: a.cfg.RuntimeID, Health: "stopped"}, nil
	case <-ctx.Done():
		return Status{}, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.status, nil
	case <-a.stoppedCh:
		return Status{RuntimeID: a.cfg.RuntimeID, Health: "stopped"}, nil
	case <-ctx.Done():
		return Status{}, ctx.Err()
	}
}
