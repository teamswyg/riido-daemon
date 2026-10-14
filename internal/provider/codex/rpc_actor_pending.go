package codex

import "errors"

func closeRPCActorPending(pending map[int64]chan RPCResult) {
	for id, ch := range pending {
		ch <- RPCResult{Err: errors.New("rpc actor closed")}
		delete(pending, id)
	}
}

func resolveRPCActorPending(pending map[int64]chan RPCResult, msg resolveMsg) {
	ch, ok := pending[msg.id]
	if !ok {
		return
	}
	delete(pending, msg.id)
	ch <- RPCResult{Result: msg.result, Err: msg.err}
}
