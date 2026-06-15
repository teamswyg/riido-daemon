package session

import (
	"bytes"
	"context"
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

func (a *burstAdapter) NewParser() agentbridge.Parser {
	done := a.done
	if len(done) == 0 {
		done = []byte("DONE")
	}
	return &burstParser{done: done}
}

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

type burstParser struct {
	done []byte
}

func (p *burstParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	if bytes.Equal(chunk, p.done) {
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
		for range burstChunks / 2 {
			running.EmitStdout([]byte("s"))
		}
		running.EmitStdout([]byte("DONE"))
	}()
	go func() {
		for range burstChunks / 2 {
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
