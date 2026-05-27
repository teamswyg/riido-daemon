package codex

import (
	"context"
	"errors"
)

// RPCResult is what a pending Register call eventually receives.
type RPCResult struct {
	Result map[string]any
	Err    error
}

// RPCActor owns the JSON-RPC pending-request map. It is the SOLE
// goroutine that touches that map; all callers interact via channels.
// This keeps the pending map actor-owned and mutex-free.
type RPCActor struct {
	idCh       chan chan int64
	registerCh chan registerMsg
	resolveCh  chan resolveMsg
	closeCh    chan struct{}
	closedCh   chan struct{}
}

type registerMsg struct {
	id    int64
	reply chan RPCResult
}

type resolveMsg struct {
	id     int64
	result map[string]any
	err    error
}

// StartRPCActor launches the actor goroutine and returns a handle.
// The actor runs until Close() is called or ctx is canceled.
func StartRPCActor(ctx context.Context) *RPCActor {
	a := &RPCActor{
		idCh:       make(chan chan int64),
		registerCh: make(chan registerMsg, 16),
		resolveCh:  make(chan resolveMsg, 16),
		closeCh:    make(chan struct{}),
		closedCh:   make(chan struct{}),
	}
	go a.run(ctx)
	return a
}

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

func (a *RPCActor) run(ctx context.Context) {
	defer close(a.closedCh)

	var nextID int64 = 1
	pending := map[int64]chan RPCResult{}

	cleanup := func() {
		for id, ch := range pending {
			ch <- RPCResult{Err: errors.New("rpc actor closed")}
			delete(pending, id)
		}
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
			ch, ok := pending[msg.id]
			if !ok {
				continue
			}
			delete(pending, msg.id)
			ch <- RPCResult{Result: msg.result, Err: msg.err}
		}
	}
}
