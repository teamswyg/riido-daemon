package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// fakeDriver is the test scaffolding ProtocolDriver. Each hook is
// configurable; all hooks count their invocations so tests can assert
// on lifecycle ordering.
//
// State is touched only by the session actor's goroutine (the actor
// calls every hook in sequence), so no mutex is required — except for
// the test reading "stdin history" while the driver may still write.
// We collect stdin writes into a channel instead, keeping the audit's
// no-mutex discipline.
type fakeDriver struct {
	startStdin []byte // frame to write in OnStart

	onStart       func(ctx context.Context, io ProtocolIO) error
	onRaw         func(ctx context.Context, raw agentbridge.RawEvent, io ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error)
	onProcessExit func(ctx context.Context, status agentbridge.ProcessExitStatus, io ProtocolIO) ([]agentbridge.Event, error)
	onClose       func(ctx context.Context, io ProtocolIO) error

	// Invocation counters — read after session terminates.
	startCalls int
	rawCalls   int
	exitCalls  int
	closeCalls int
}

func (d *fakeDriver) OnStart(ctx context.Context, io ProtocolIO) error {
	d.startCalls++
	if d.onStart != nil {
		return d.onStart(ctx, io)
	}
	if len(d.startStdin) > 0 {
		return io.WriteStdin(ctx, d.startStdin)
	}
	return nil
}

func (d *fakeDriver) OnRaw(ctx context.Context, raw agentbridge.RawEvent, io ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	d.rawCalls++
	if d.onRaw != nil {
		return d.onRaw(ctx, raw, io)
	}
	return nil, nil, nil
}

func (d *fakeDriver) OnProcessExit(ctx context.Context, status agentbridge.ProcessExitStatus, io ProtocolIO) ([]agentbridge.Event, error) {
	d.exitCalls++
	if d.onProcessExit != nil {
		return d.onProcessExit(ctx, status, io)
	}
	return nil, nil
}

func (d *fakeDriver) OnClose(ctx context.Context, io ProtocolIO) error {
	d.closeCalls++
	if d.onClose != nil {
		return d.onClose(ctx, io)
	}
	return nil
}

// trackingAdapter records whether Translate was called — used to prove
// the driver path BYPASSES Translate when a driver is installed.
//
// Translate runs only on the session actor's goroutine. The test reads
// translateCalls only AFTER waiting on sess.Done(), whose close is the
// last thing the actor does — that close establishes the happens-before
// edge that makes the unsynchronized field read race-free. No mutex.
type trackingAdapter struct {
	translateCalls int
}

func (a *trackingAdapter) Name() string { return "tracking" }
func (a *trackingAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}
func (a *trackingAdapter) BuildStart(_ agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{}, nil
}
func (a *trackingAdapter) NewParser() agentbridge.Parser { return &echoParser{} }
func (a *trackingAdapter) Translate(_ agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	a.translateCalls++ // session actor goroutine is the sole writer
	return nil, nil, nil
}
func (a *trackingAdapter) BlockedArgs() []string { return nil }

// TranslateCalls is safe to call only after the session has fully
// terminated (i.e. after <-sess.Done()).
func (a *trackingAdapter) TranslateCalls() int { return a.translateCalls }

// echoParser turns every chunk into one RawEvent of Type "chunk".
type echoParser struct{}

func (p *echoParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}
func (p *echoParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (p *echoParser) Close() ([]agentbridge.RawEvent, error)                  { return nil, nil }

// --- A1. OnStart is called before the loop runs, and its stdin writes
//     land in the process's stdin history. ---

func TestSessionCallsProtocolDriverOnStart(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	driver := &fakeDriver{startStdin: []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-onstart",
		RuntimeID:      "rt-1",
		Adapter:        &trackingAdapter{},
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	running := sess.runningForTest()

	// The first stdin frame should be the OnStart write.
	select {
	case b := <-running.StdinRecv():
		if !bytesContain(b, "initialize") {
			t.Fatalf("OnStart stdin: %q", b)
		}
	case <-time.After(time.Second):
		t.Fatal("OnStart stdin frame never written")
	}

	// Tear down. Drain events so the actor doesn't block on send, then
	// wait on Done() — closed AFTER OnClose has run, so reading
	// driver.startCalls / closeCalls is race-free.
	go func() {
		for range sess.Events() {
		}
	}()
	running.EmitExit(0, nil)
	<-sess.Result()
	<-sess.Done()
	if driver.startCalls != 1 {
		t.Fatalf("OnStart calls: %d", driver.startCalls)
	}
	if driver.closeCalls != 1 {
		t.Fatalf("OnClose calls: %d", driver.closeCalls)
	}
}

// --- A2. Raw events route through driver.OnRaw — adapter.Translate is
//     NOT called. ---

func TestSessionRoutesRawThroughProtocolDriver(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	tracker := &trackingAdapter{}
	driver := &fakeDriver{
		onRaw: func(_ context.Context, raw agentbridge.RawEvent, _ ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type != "chunk" {
				return nil, nil, nil
			}
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "driver-done"}}}, nil, nil
			}
			return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}}, nil, nil
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-onraw",
		Adapter:        tracker,
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	running := sess.runningForTest()
	go func() {
		running.EmitStdout([]byte("hello"))
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted || res.Output != "driver-done" {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no result")
	}

	// Wait for the actor's last write to be fully published before
	// reading the unsynchronized counter — Done() closes last.
	<-sess.Done()
	if tracker.TranslateCalls() != 0 {
		t.Fatalf("Translate must NOT be called when driver is installed, got %d calls", tracker.TranslateCalls())
	}
	if driver.rawCalls == 0 {
		t.Fatalf("OnRaw was never called")
	}
}

// --- A3. OnProcessExit is invoked and its events feed the reducer. ---

func TestSessionProtocolDriverProcessExitCleansUp(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	driver := &fakeDriver{
		onProcessExit: func(_ context.Context, status agentbridge.ProcessExitStatus, _ ProtocolIO) ([]agentbridge.Event, error) {
			// Simulate "fail every pending RPC request" — emit one
			// Error event the reducer will pass through (logged).
			return []agentbridge.Event{{
				Kind: agentbridge.EventError,
				Err:  "driver: 1 pending request cancelled due to process exit code " + itoa(status.Code),
			}}, nil
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-exit",
		Adapter:        &trackingAdapter{},
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	running := sess.runningForTest()

	// Drain events.
	gotDriverError := make(chan struct{}, 1)
	go func() {
		for ev := range sess.Events() {
			if ev.Kind == agentbridge.EventError && bytesContain([]byte(ev.Err), "driver:") {
				select {
				case gotDriverError <- struct{}{}:
				default:
				}
			}
		}
	}()

	running.EmitExit(2, nil)

	select {
	case <-gotDriverError:
	case <-time.After(2 * time.Second):
		t.Fatal("driver-emitted error event never reached events stream")
	}
	<-sess.Result()
	if driver.exitCalls != 1 {
		t.Fatalf("OnProcessExit calls: %d", driver.exitCalls)
	}
}

// --- A4. nil ProtocolDriver keeps the legacy Adapter.Translate path. ---

func TestSessionWithoutProtocolDriverKeepsLegacyPath(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	tracker := &trackingAdapter{}
	sess, err := Start(context.Background(), Config{
		TaskID:  "t-legacy",
		Adapter: tracker,
		Process: fake,
		Spawn:   process.Command{Executable: "fake"},
		// ProtocolDriver intentionally nil.
	})
	if err != nil {
		t.Fatal(err)
	}
	running := sess.runningForTest()
	go func() {
		running.EmitStdout([]byte("x"))
		running.EmitExit(0, nil)
	}()
	<-sess.Result()
	<-sess.Done()
	if tracker.TranslateCalls() == 0 {
		t.Fatal("Translate must be called on the legacy path")
	}
}

// --- A5. OnStart returning error terminates the session with Failed. ---

func TestSessionProtocolDriverOnStartErrorTerminatesFailed(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	driver := &fakeDriver{
		onStart: func(_ context.Context, _ ProtocolIO) error {
			return errors.New("init handshake failed")
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:         "t-onstart-err",
		Adapter:        &trackingAdapter{},
		Process:        fake,
		Spawn:          process.Command{Executable: "fake"},
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Drain events so the actor doesn't block.
	go func() {
		for range sess.Events() {
		}
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultFailed {
			t.Fatalf("status: %s", res.Status)
		}
		if res.Error == "" || !bytesContain([]byte(res.Error), "init handshake failed") {
			t.Fatalf("result error: %q", res.Error)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("session never terminated on OnStart error")
	}

	// Kill the fake process if it's still alive so test cleanup doesn't leak.
	running := sess.runningForTest()
	func() {
		defer func() { _ = recover() }()
		running.EmitExit(0, nil)
	}()
}

// --- helpers ---

func bytesContain(b []byte, sub string) bool {
	if len(sub) > len(b) {
		return false
	}
	for i := 0; i+len(sub) <= len(b); i++ {
		if string(b[i:i+len(sub)]) == sub {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
