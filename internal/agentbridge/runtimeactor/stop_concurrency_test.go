package runtimeactor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
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
	for i := range N {
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
	for trial := range 20 {
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

func TestStopLifecycleForcedEscalatesGracefulDrain(t *testing.T) {
	proc := newBlockingKillProcess()
	t.Cleanup(proc.unblock)

	a, err := New(Config{
		RuntimeID: "rt-stop-escalate",
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		Process:       proc,
		MaxConcurrent: 1,
		MailboxSize:   8,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := a.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownForced, time.Second)
		defer cancel()
		_ = a.StopLifecycle(ctx)
	})

	submitCtx, submitCancel := context.WithTimeout(context.Background(), time.Second)
	defer submitCancel()
	if _, err := a.Submit(submitCtx, bridge.TaskRequest{ID: "t-stuck", Provider: "fake"}); err != nil {
		t.Fatalf("Submit: %v", err)
	}

	graceCtx, graceCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer graceCancel()
	if err := a.Stop(graceCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("graceful Stop error = %v, want context deadline exceeded", err)
	}
	select {
	case <-proc.running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("graceful Stop did not request provider kill")
	}

	forcedCtx, forcedCancel := lifecycle.DetachedShutdown(lifecycle.ShutdownForced, time.Second)
	defer forcedCancel()
	if err := a.StopLifecycle(forcedCtx); err != nil {
		t.Fatalf("forced StopLifecycle: %v", err)
	}
}

type blockingKillProcess struct {
	running *blockingKillRunning
}

func newBlockingKillProcess() *blockingKillProcess {
	return &blockingKillProcess{running: newBlockingKillRunning()}
}

func (p *blockingKillProcess) Start(_ context.Context, _ process.Command) (process.RunningProcess, error) {
	return p.running, nil
}

func (p *blockingKillProcess) unblock() {
	close(p.running.unblock)
}
