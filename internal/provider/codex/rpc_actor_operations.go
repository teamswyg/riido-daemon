package codex

import "errors"

// NextID returns the next monotonic JSON-RPC id. Safe from any goroutine.
func (a *RPCActor) NextID() int64 {
	reply := make(chan int64, 1)
	select {
	case a.idCh <- reply:
	case <-a.closedCh:
		return 0
	}
	select {
	case id := <-reply:
		return id
	case <-a.closedCh:
		return 0
	}
}

// Register reserves a slot for id and returns a channel on which the
// caller will receive the RPCResult once Resolve(id, ...) is called.
// The channel has capacity 1 so Resolve never blocks even if the caller
// has not yet started reading.
func (a *RPCActor) Register(id int64) <-chan RPCResult {
	reply := make(chan RPCResult, 1)
	select {
	case a.registerCh <- registerMsg{id: id, reply: reply}:
	case <-a.closedCh:
		reply <- RPCResult{Err: errors.New("rpc actor closed")}
	}
	return reply
}

// Resolve delivers the result for id. If id was never Register'd, the
// call is a no-op (does not panic, does not block). If both result and
// err are nil, an empty result is delivered.
func (a *RPCActor) Resolve(id int64, result map[string]any, err error) {
	select {
	case a.resolveCh <- resolveMsg{id: id, result: result, err: err}:
	case <-a.closedCh:
	}
}

// Close stops the actor. Pending Register callers receive an error
// result so they don't leak. Safe to call multiple times.
func (a *RPCActor) Close() {
	select {
	case <-a.closeCh:
		return
	default:
		close(a.closeCh)
	}
	<-a.closedCh
}
