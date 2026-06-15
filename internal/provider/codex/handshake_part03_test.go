package codex

import (
	"context"
	"strconv"
	"testing"
	"time"
)

// TestCodexRPCActorReleasesPendingOnProcessExit is the M-3 invariant
// "process exit with pending requests" — pending callers must NOT leak
// when the RPC actor closes for any reason, including because the
// process backing it died.
func TestCodexRPCActorReleasesPendingOnProcessExit(t *testing.T) {
	rpc := StartRPCActor(context.Background())

	const N = 8
	replies := make([]<-chan RPCResult, 0, N)
	for range N {
		id := rpc.NextID()
		replies = append(replies, rpc.Register(id))
	}

	// Simulate process exit: caller closes the RPC actor.
	rpc.Close()

	deadline := time.After(2 * time.Second)
	for i, ch := range replies {
		select {
		case r := <-ch:
			if r.Err == nil {
				t.Fatalf("reply #%d: expected error after Close, got nil", i)
			}
		case <-deadline:
			t.Fatalf("reply #%d: blocked after Close", i)
		}
	}
	_ = strconv.Itoa // keep import set stable
}
