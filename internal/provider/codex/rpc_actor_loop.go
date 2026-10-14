package codex

import "context"

func (a *RPCActor) run(ctx context.Context) {
	defer close(a.closedCh)

	var nextID int64 = 1
	pending := map[int64]chan RPCResult{}
	cleanup := func() {
		closeRPCActorPending(pending)
	}

	for {
		select {
		case <-ctx.Done():
			cleanup()
			return
		case <-a.closeCh:
			cleanup()
			return
		case reply := <-a.idCh:
			reply <- nextID
			nextID++
		case msg := <-a.registerCh:
			pending[msg.id] = msg.reply
		case msg := <-a.resolveCh:
			resolveRPCActorPending(pending, msg)
		}
	}
}
