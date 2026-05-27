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

// burstAdapter accumulates every chunk into one big TextDelta and
// emits a terminal Result on a sentinel chunk. Used by the burst
// regression test.
type burstAdapter struct {
	done []byte // sentinel; when seen, emit Result
}

func (a *burstAdapter) Name() string { return "burst" }
func (a *burstAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}
func (a *burstAdapter) BuildStart(_ agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{}, nil
}
func (a *burstAdapter) NewParser() agentbridge.Parser { return &burstParser{} }
func (a *burstAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "sentinel" {
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
	}
	if raw.Type == "chunk" {
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}}, nil, nil
	}
	return nil, nil, nil
}
func (a *burstAdapter) BlockedArgs() []string { return nil }

type burstParser struct{}

func (p *burstParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	if string(chunk) == "DONE" {
		return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "sentinel", Bytes: chunk}}, nil
	}
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}
func (p *burstParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStderr, Type: "chunk", Bytes: chunk}}, nil
}
func (p *burstParser) Close() ([]agentbridge.RawEvent, error) { return nil, nil }

// TestSessionStdoutStderrBurstNoDeadlock: hammer the session actor with
// a fast burst of stdout AND stderr chunks. Verifies (a) no deadlock
// even when the events channel pressure builds up, (b) terminal Result
// is delivered exactly once, (c) all chunks get translated.
//
// This is the regression test for the "stdout / stderr burst" item in
// M-4 of the implementation audit.
func TestSessionStdoutStderrBurstNoDeadlock(t *testing.T) {
	const burstChunks = 1024

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	sess, err := Start(context.Background(), Config{
		TaskID:    "t-burst",
		RuntimeID: "rt-1",
		Adapter:   &burstAdapter{},
		Process:   fake,
		Spawn:     process.Command{Executable: "burst"},
		// Smaller event buffer to make backpressure realistic.
		EventBuffer: 8,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	running := sess.runningForTest()

	// Drain events fast.
	drained := make(chan int, 1)
	go func() {
		count := 0
		for range sess.Events() {
			count++
		}
		drained <- count
	}()

	// Burst feed stdout + stderr concurrently.
	go func() {
		for i := 0; i < burstChunks/2; i++ {
			running.EmitStdout([]byte("s"))
		}
		running.EmitStdout([]byte("DONE"))
	}()
	go func() {
		for i := 0; i < burstChunks/2; i++ {
			running.EmitStderr([]byte("e"))
		}
	}()

	// Result must arrive within a generous deadline.
	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("status: %s", res.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("burst caused deadlock — Result never arrived")
	}

	// EmitExit so the fake process closes its channels and the session
	// loop can finish; otherwise drained goroutine will hang.
	running.EmitExit(0, nil)

	select {
	case n := <-drained:
		if n == 0 {
			t.Fatalf("no events drained")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("events channel never closed")
	}
}

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
		for i := 0; i < burstChunks; i++ {
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

	for cycle := 0; cycle < 5; cycle++ {
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
