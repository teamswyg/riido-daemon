package runtimeactor

import "context"

// HeartbeatPayload returns the publish-ready heartbeat.
func (a *Actor) HeartbeatPayload(ctx context.Context) (Heartbeat, error) {
	reply := make(chan statusReply, 1)
	select {
	case a.statusCh <- statusMsg{ctx: ctx, reply: reply}:
	case <-a.stoppedCh:
		return Heartbeat{RuntimeID: a.cfg.RuntimeID}, nil
	case <-ctx.Done():
		return Heartbeat{}, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.hb, nil
	case <-a.stoppedCh:
		return Heartbeat{RuntimeID: a.cfg.RuntimeID}, nil
	case <-ctx.Done():
		return Heartbeat{}, ctx.Err()
	}
}
