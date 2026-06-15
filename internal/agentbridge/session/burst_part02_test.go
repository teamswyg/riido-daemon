package session

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// TestSessionCancelDuringBurstDoesNotDeadlock: cancel arrives in the
// middle of a heavy burst. Result must still be delivered (as
// ResultCancelled) and the events channel must close.
func TestSessionCancelDuringBurstDoesNotDeadlock(t *testing.T) {
	const burstChunks = 2048

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	sess, _ := Start(context.Background(), Config{
		TaskID:      "t-burst-cancel",
		RuntimeID:   "rt-1",
		Adapter:     &burstAdapter{},
		Process:     fake,
		Spawn:       process.Command{Executable: "burst"},
		EventBuffer: 8,
	})
	running := sess.runningForTest()

	// Drain events.
	drained := make(chan struct{})
	go func() {
		for range sess.Events() {
		}
		close(drained)
	}()

	burstDone := make(chan struct{})
	go func() {
		defer close(burstDone)
		for range burstChunks {
			// Allow EmitStdout to recover from a closed channel after
			// cancellation by guarding with recover().
			func() {
				defer func() { _ = recover() }()
				running.EmitStdout([]byte("x"))
			}()
		}
	}()

	// Wait a tick so some chunks are in flight, then cancel.
	time.Sleep(20 * time.Millisecond)
	sess.Cancel(nil)

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("status: %s", res.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancel-during-burst caused deadlock")
	}

	// Final fake teardown so the burst goroutine can finish if still running.
	func() {
		defer func() { _ = recover() }()
		running.EmitExit(0, nil)
	}()

	select {
	case <-drained:
	case <-time.After(2 * time.Second):
		t.Fatal("events channel didn't close after cancel")
	}
	<-burstDone
}

// TestSessionGoroutineCleanupAfterCompletion is the session-level leak
// guard. After Start → drive to Result → wait, NumGoroutine should
// return to (baseline + small tolerance).
func TestSessionGoroutineCleanupAfterCompletion(t *testing.T) {
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	for cycle := range 5 {
		fake := process.NewFake()
		fake.NextRunning = process.NewFakeRunning()
		sess, err := Start(context.Background(), Config{
			TaskID:    "t-" + strings.Repeat("x", cycle+1),
			RuntimeID: "rt-1",
			Adapter:   &burstAdapter{},
			Process:   fake,
			Spawn:     process.Command{Executable: "x"},
		})
		if err != nil {
			t.Fatal(err)
		}
		running := sess.runningForTest()

		// Drain events.
		go func() {
			for range sess.Events() {
			}
		}()
		go func() {
			running.EmitStdout([]byte("DONE"))
			running.EmitExit(0, nil)
		}()
		<-sess.Result()
	}

	final := waitNumGoroutine(baseline+2, 2*time.Second)
	if final > baseline+2 {
		t.Fatalf("session goroutine leak: baseline=%d final=%d", baseline, final)
	}
}

func waitNumGoroutine(target int, deadline time.Duration) int {
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		if runtime.NumGoroutine() <= target {
			return runtime.NumGoroutine()
		}
		time.Sleep(20 * time.Millisecond)
	}
	return runtime.NumGoroutine()
}
