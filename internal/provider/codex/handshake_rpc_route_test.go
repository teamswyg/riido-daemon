package codex

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func routeRPC(t *testing.T, parser agentbridge.Parser, stdout <-chan []byte, rpc *RPCActor, deadline time.Duration, awaitID int64) {
	t.Helper()
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		select {
		case chunk, ok := <-stdout:
			if !ok {
				return
			}
			if routeRPCChunk(parser, rpc, chunk, awaitID) {
				return
			}
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func routeRPCChunk(parser agentbridge.Parser, rpc *RPCActor, chunk []byte, awaitID int64) bool {
	raws, _ := parser.FeedStdout(chunk)
	for _, raw := range raws {
		if raw.Type != "response" {
			continue
		}
		id, hasID := rpcID(raw.Payload)
		if !hasID {
			continue
		}
		rpc.Resolve(id, mapField(raw.Payload, "result"), nil)
		if id == awaitID {
			return true
		}
	}
	return false
}

func routeHandshakeRawRPC(raw agentbridge.RawEvent, rpc *RPCActor) {
	switch raw.Type {
	case "response":
		if id, ok := rpcID(raw.Payload); ok {
			rpc.Resolve(id, mapField(raw.Payload, "result"), nil)
		}
	case "error":
		if id, ok := rpcID(raw.Payload); ok {
			rpc.Resolve(id, nil, jsonRPCError(raw.Payload))
		}
	}
}

func jsonRPCError(p map[string]any) error {
	e, _ := p["error"].(map[string]any)
	msg, _ := e["message"].(string)
	if msg == "" {
		msg = "unknown rpc error"
	}
	return &rpcErr{msg: msg}
}

type rpcErr struct{ msg string }

func (e *rpcErr) Error() string { return e.msg }
