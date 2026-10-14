package codex

import "context"

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
