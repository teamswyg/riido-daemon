package codex

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func drainHandshake(
	t *testing.T,
	parser agentbridge.Parser,
	stdout <-chan []byte,
	rpc *RPCActor,
	deadline time.Duration,
	predicate func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool,
) {
	t.Helper()
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		select {
		case chunk, ok := <-stdout:
			if !ok {
				return
			}
			if drainHandshakeChunk(t, parser, rpc, predicate, chunk) {
				return
			}
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func drainHandshakeChunk(
	t *testing.T,
	parser agentbridge.Parser,
	rpc *RPCActor,
	predicate func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool,
	chunk []byte,
) bool {
	t.Helper()
	raws, err := parser.FeedStdout(chunk)
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	for _, raw := range raws {
		routeHandshakeRawRPC(raw, rpc)
		evs, _, err := Translate(raw)
		if err != nil {
			t.Fatalf("translate: %v", err)
		}
		if predicate(raw, evs) {
			return true
		}
	}
	return false
}
