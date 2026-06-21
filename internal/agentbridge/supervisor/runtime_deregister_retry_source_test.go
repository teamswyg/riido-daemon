package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"
)

type runtimeDeregisterRetrySource struct {
	*runtimeRoutingSource
	failures int
	attempts chan int
	count    int
}

func newRuntimeDeregisterRetrySource(failures int) *runtimeDeregisterRetrySource {
	return &runtimeDeregisterRetrySource{
		runtimeRoutingSource: newRuntimeRoutingSource(nil),
		failures:             failures,
		attempts:             make(chan int, failures+2),
	}
}

func (s *runtimeDeregisterRetrySource) DeregisterRuntime(ctx context.Context, runtimeID string) error {
	s.count++
	s.attempts <- s.count
	if s.count <= s.failures {
		return errors.New("deregister rejected")
	}
	return s.runtimeRoutingSource.DeregisterRuntime(ctx, runtimeID)
}

func expectDeregisterAttempt(t *testing.T, source *runtimeDeregisterRetrySource, want int) {
	t.Helper()
	select {
	case got := <-source.attempts:
		if got != want {
			t.Fatalf("deregister attempt = %d, want %d", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("deregister attempt %d was not observed", want)
	}
}
