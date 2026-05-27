package runtimeactor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// TestStopIsIdempotentSerially: calling Stop a second time after the
// actor has already fully stopped must return nil immediately. The
// previous close(stopCh)+recover() pattern relied on a deferred
// recover to absorb the second close; the stopReqCh pattern absorbs
// it via a non-blocking send into a capacity-1 buffer.
func TestStopIsIdempotentSerially(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := a.Stop(ctx); err != nil {
		t.Fatalf("first stop: %v", err)
	}
	// Second stop must NOT panic and must return nil.
	if err := a.Stop(ctx); err != nil {
		t.Fatalf("second stop: %v", err)
	}
	// Third stop too.
	if err := a.Stop(ctx); err != nil {
		t.Fatalf("third stop: %v", err)
	}
}

// TestStopIsIdempotentConcurrent: N goroutines call Stop simultaneously.
// With the previous close(stopCh) approach this would have panicked on
// multiple concurrent close() calls without the recover guard. The
// new stopReqCh send-only pattern admits no panic path at all.
func TestStopIsIdempotentConcurrent(t *testing.T) {
	a, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})

	const N = 8
	var wg sync.WaitGroup
	wg.Add(N)
	errs := make([]error, N)
	start := make(chan struct{})
	for i := 0; i < N; i++ {
		go func(idx int) {
			defer wg.Done()
			<-start // release everyone at once
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			errs[idx] = a.Stop(ctx)
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("Stop goroutine %d returned error: %v", i, err)
		}
	}
}

// TestStopWhileSubmitInFlight: Submit and Stop race. The combination
// must always terminate — neither call may deadlock. Submit may
// succeed-and-then-be-cancelled OR return ErrActorStopped; we accept
// either outcome and only assert "did not hang."
func TestStopWhileSubmitInFlight(t *testing.T) {
	for trial := 0; trial < 20; trial++ {
		a, _ := startActor(t, Config{
			Adapters: []agentbridge.Adapter{
				&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
			},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

		submitDone := make(chan struct{})
		go func() {
			defer close(submitDone)
			_, _ = a.Submit(ctx, bridge.TaskRequest{ID: "t-race", Provider: "fake"})
		}()

		stopDone := make(chan struct{})
		go func() {
			defer close(stopDone)
			_ = a.Stop(ctx)
		}()

		select {
		case <-submitDone:
		case <-time.After(3 * time.Second):
			cancel()
			t.Fatalf("trial %d: Submit blocked indefinitely", trial)
		}
		select {
		case <-stopDone:
		case <-time.After(3 * time.Second):
			cancel()
			t.Fatalf("trial %d: Stop blocked indefinitely", trial)
		}
		cancel()
	}
}
