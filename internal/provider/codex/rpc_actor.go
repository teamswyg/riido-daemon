package codex

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
