package codex

import "time"

func (f *codexHandshakeFixture) initialize() {
	initID := f.rpc.NextID()
	initReply := f.rpc.Register(initID)
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0",
		"id":      initID,
		"method":  "initialize",
		"params":  map[string]any{"client": "riido"},
	})
	f.running.EmitStdout(jsonline(f.t, map[string]any{
		"jsonrpc": "2.0",
		"id":      initID,
		"result":  map[string]any{"serverInfo": map[string]any{"name": "codex"}},
	}))
	routeRPC(f.t, f.parser, f.running.Stdout(), f.rpc, time.Second, initID)
	select {
	case result := <-initReply:
		if result.Err != nil {
			f.t.Fatalf("initialize err: %v", result.Err)
		}
		if result.Result["serverInfo"] == nil {
			f.t.Fatalf("initialize result missing serverInfo: %+v", result.Result)
		}
	case <-time.After(time.Second):
		f.t.Fatal("initialize reply not delivered")
	}
}

func (f *codexHandshakeFixture) notifyInitialized() {
	mustWriteJSONRPC(f.t, f.running, map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]any{},
	})
}
