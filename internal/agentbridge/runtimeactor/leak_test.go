package runtimeactor

import (
	"context"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// waitGoroutineCount polls runtime.NumGoroutine until it equals target
// (or deadline). Test scheduling is non-deterministic; a short polling
// window absorbs goroutine teardown latency without making the test
// flake. Returns the final count for diagnostic messages.
func waitGoroutineCount(target int, deadline time.Duration) int {
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		if runtime.NumGoroutine() <= target {
			return runtime.NumGoroutine()
		}
		time.Sleep(20 * time.Millisecond)
	}
	return runtime.NumGoroutine()
}

// TestRuntimeActorStopReleasesAllGoroutines is the load-bearing leak
// test for M-4. After Submit→Complete→Stop the NumGoroutine count must
// return to within a small tolerance of the baseline captured before
// the actor was created.
//
// Tolerance: we allow +2 because Go's runtime keeps a few worker
// goroutines that can spawn and exit asynchronously around scheduler
// activity. Anything beyond that indicates a real leak.
func TestRuntimeActorStopReleasesAllGoroutines(t *testing.T) {
	// Force a GC + warm-up so transient test goroutines have settled.
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	for cycle := range 3 {
		proc := newFakeProcess()
		a, err := New(Config{
			RuntimeID: "rt-leak",
			Adapters: []agentbridge.Adapter{
				&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
			},
			Process:       proc,
			MaxConcurrent: 2,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := a.Start(context.Background()); err != nil {
			t.Fatal(err)
		}

		// Run a couple of tasks per cycle.
		for i := range 2 {
			h, err := a.Submit(context.Background(), bridge.TaskRequest{
				ID: "t-" + strconv.Itoa(cycle) + "-" + strconv.Itoa(i), Provider: "fake",
			})
			if err != nil {
				t.Fatal(err)
			}
			r := waitForRunning(t, proc, i, time.Second)
			go func() {
				r.EmitStdout([]byte("done"))
				r.EmitExit(0, nil)
			}()
			<-h.Result()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := a.Stop(ctx); err != nil {
			t.Fatalf("Stop cycle %d: %v", cycle, err)
		}
		cancel()
	}

	// Allow stragglers (e.g. the fakeProcess actor goroutines) up to a
	// second to exit. Each fakeProcess we created is leaked-by-design
	// in the cycle (we don't expose a Stop on it). So we expect at
	// most 3 extra goroutines (one per fakeProcess across cycles).
	final := waitGoroutineCount(baseline+3, 2*time.Second)
	if final > baseline+3 {
		t.Fatalf("goroutine leak: baseline=%d final=%d (delta=%d)", baseline, final, final-baseline)
	}
}

// TestRuntimeActorCancelCascadesToProcessAndSession verifies the full
// chain: Runtime.Cancel → Session.Cancel → Process.Kill → Result is
// ResultCancelled, and the slot is freed afterwards.
func TestRuntimeActorCancelCascadesToProcessAndSession(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})

	h, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-cascade", Provider: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	r := waitForRunning(t, p, 0, time.Second)

	if err := a.Cancel(context.Background(), "t-cascade", "test cascade"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	// 1. Process must receive Kill.
	select {
	case <-r.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("process kill never received")
	}

	// 2. Result must be ResultCancelled.
	select {
	case res := <-h.Result():
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("status: %s", res.Status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("session never terminated")
	}

	// 3. Slot must free up.
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		s, _ := a.Status(context.Background())
		if s.RunningSessions == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("slot never released after cascade-cancel")
}

// TestRuntimeActorStopCascadesToInFlightSessions: when Stop is called
// with running sessions, each session's process receives a Kill signal
// and each session terminates with ResultCancelled. Already covered
// by TestRuntimeActorShutdownCancelsRunningSessions in runtimeactor_test.go,
// but we add a stronger check that the FakeRunning's KillRecv fires
// for each session (not just session.Cancel).
func TestRuntimeActorStopCascadesToProcesses(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
		MaxConcurrent: 3,
	})

	for i := range 3 {
		_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "t-" + strconv.Itoa(i), Provider: "fake"})
		if err != nil {
			t.Fatal(err)
		}
		_ = waitForRunning(t, p, i, time.Second)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = a.Stop(ctx)
	}()

	for i := range 3 {
		r := p.at(i)
		if r == nil {
			t.Fatalf("running #%d missing", i)
		}
		select {
		case <-r.KillRecv():
		case <-time.After(3 * time.Second):
			t.Fatalf("process #%d kill never received", i)
		}
	}
}
