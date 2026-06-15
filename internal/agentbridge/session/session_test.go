package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// recordingAdapter is a minimal Adapter that fakes Translate by mapping
// raw Type onto pre-canned events. It lets the session-actor tests
// exercise the wiring without depending on a real provider.
type recordingAdapter struct {
	name        string
	startCmd    agentbridge.StartCommand
	blocked     []string
	translateFn func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error)
	inputFn     func(agentbridge.Command) ([]byte, error)
	parser      *recordingParser
}

func (a *recordingAdapter) Name() string { return a.name }
func (a *recordingAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (a *recordingAdapter) BuildStart(_ agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return a.startCmd, nil
}
func (a *recordingAdapter) NewParser() agentbridge.Parser { return a.parser }
func (a *recordingAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if a.translateFn != nil {
		return a.translateFn(raw)
	}
	return nil, nil, nil
}

func (a *recordingAdapter) BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	if a.inputFn != nil {
		return a.inputFn(cmd)
	}
	return nil, nil
}
func (a *recordingAdapter) BlockedArgs() []string { return a.blocked }

// recordingParser stores raw chunks for inspection and emits one RawEvent
// of Type "chunk" per FeedStdout call, so the wiring is trivially testable.
type recordingParser struct {
	stdoutChunks [][]byte
	stderrChunks [][]byte
	closed       bool
}

func (p *recordingParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.stdoutChunks = append(p.stdoutChunks, append([]byte(nil), chunk...))
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}

func (p *recordingParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.stderrChunks = append(p.stderrChunks, append([]byte(nil), chunk...))
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStderr, Type: "stderr-chunk", Bytes: chunk}}, nil
}

func (p *recordingParser) Close() ([]agentbridge.RawEvent, error) {
	p.closed = true
	return nil, nil
}

// drainEvents reads from the session's events channel until it is closed
// or the deadline ticks.
func drainEvents(t *testing.T, sess *Session, deadline time.Duration) []agentbridge.Event {
	t.Helper()
	var out []agentbridge.Event
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	for {
		select {
		case ev, ok := <-sess.Events():
			if !ok {
				return out
			}
			out = append(out, ev)
		case <-timer.C:
			t.Fatal("drainEvents deadline exceeded")
			return out
		}
	}
}

// waitResult reads the terminal Result with a deadline.
func waitResult(t *testing.T, sess *Session, deadline time.Duration) agentbridge.Result {
	t.Helper()
	select {
	case res := <-sess.Result():
		return res
	case <-time.After(deadline):
		t.Fatal("waitResult deadline exceeded")
		return agentbridge.Result{}
	}
}

func TestSessionDefaultBuffersMatchProviderRuntimeBackpressureSSOT(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			}
			return nil, nil, nil
		},
	}

	sess, err := Start(context.Background(), Config{
		TaskID:    "task-default-buffer",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := cap(sess.Events()); got != DefaultEventBuffer {
		t.Fatalf("default event buffer = %d, want %d", got, DefaultEventBuffer)
	}
	if got := cap(sess.Result()); got != DefaultResultBuffer {
		t.Fatalf("default result buffer = %d, want %d", got, DefaultResultBuffer)
	}

	go func() {
		sess.runningForTest().EmitStdout([]byte("DONE"))
		sess.runningForTest().EmitExit(0, nil)
	}()
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	_ = drainEvents(t, sess, time.Second)
}
