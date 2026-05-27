package session

import (
	"context"
	"errors"
	"os"
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

// Normal completion: stdout chunk → translated TextDelta → Result.
func TestSessionNormalCompletion(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	parser := &recordingParser{}
	adapter := &recordingAdapter{
		name:   "fake",
		parser: parser,
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type != "chunk" {
				return nil, nil, nil
			}
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "hello"}}}, nil, nil
			}
			return []agentbridge.Event{
				{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning},
				{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)},
			}, nil, nil
		},
	}

	cfg := Config{
		TaskID:    "task-1",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
	}
	sess, err := Start(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	running := sess.runningForTest()

	// Drive output: a normal chunk, then DONE, then exit.
	go func() {
		running.EmitStdout([]byte("hello"))
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()

	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("status: %s", res.Status)
	}
	if res.Output != "hello" {
		t.Fatalf("output: %q", res.Output)
	}

	// After Result, Events channel must be closed.
	events := drainEvents(t, sess, time.Second)
	sawText := false
	for _, ev := range events {
		if ev.Kind == agentbridge.EventTextDelta && ev.Text == "hello" {
			sawText = true
		}
	}
	if !sawText {
		t.Fatalf("expected to see TextDelta in event stream, got %+v", events)
	}
}

func TestSessionExtractsRiidoTelemetryAboveAdapters(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			}
			return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}}, nil, nil
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:    "task-telemetry",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	running := sess.runningForTest()
	go func() {
		running.EmitStdout([]byte("<riido_log>프로젝트 go"))
		running.EmitStdout([]byte(".mod 작성중<end>"))
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	events := drainEvents(t, sess, time.Second)
	for _, ev := range events {
		if ev.Kind == agentbridge.EventProgress && ev.Text == "프로젝트 go.mod 작성중" {
			return
		}
	}
	t.Fatalf("missing progress event in %+v", events)
}

func TestSessionAutoApprovalWritesProviderInput(t *testing.T) {
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	parser := &recordingParser{}
	adapter := &recordingAdapter{
		name:   "fake",
		parser: parser,
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			switch string(raw.Bytes) {
			case "ASK":
				return []agentbridge.Event{{
					Kind: agentbridge.EventToolApprovalNeeded,
					Tool: agentbridge.ToolRef{ID: "tool-1", Kind: "read", ProviderRequestID: "req-1"},
				}}, nil, nil
			case "DONE":
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			default:
				return nil, nil, nil
			}
		},
		inputFn: func(cmd agentbridge.Command) ([]byte, error) {
			if cmd.Kind != agentbridge.CommandApproveTool || cmd.ToolID != "tool-1" || cmd.ProviderRequestID != "req-1" {
				t.Fatalf("unexpected provider input command: %+v", cmd)
			}
			return []byte("approve:req-1\n"), nil
		},
	}

	sess, err := Start(context.Background(), Config{
		TaskID:      "task-approval",
		RuntimeID:   "rt-1",
		Adapter:     adapter,
		Process:     fake,
		Spawn:       process.Command{Executable: "fake"},
		AutoApprove: func(tool agentbridge.ToolRef) bool { return tool.Kind == "read" },
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	running.EmitStdout([]byte("ASK"))
	select {
	case got := <-running.StdinRecv():
		if string(got) != "approve:req-1\n" {
			t.Fatalf("provider input = %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider input was not written")
	}

	go func() {
		running.EmitStdout([]byte("DONE"))
		running.EmitExit(0, nil)
	}()
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	_ = drainEvents(t, sess, time.Second)
}

func TestSessionToolStartGateBlocksStartedTool(t *testing.T) {
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	parser := &recordingParser{}
	adapter := &recordingAdapter{
		name:   "fake",
		parser: parser,
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "START" {
				return []agentbridge.Event{{
					Kind: agentbridge.EventToolCallStarted,
					Tool: agentbridge.ToolRef{ID: "tool-1", Kind: "shell", Args: map[string]string{"command": "rm -rf .git"}},
				}}, nil, nil
			}
			return nil, nil, nil
		},
	}

	sess, err := Start(context.Background(), Config{
		TaskID:    "task-tool-start-block",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
		ToolStartGate: func(tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
			if tool.ID != "tool-1" {
				t.Fatalf("unexpected tool: %+v", tool)
			}
			return agentbridge.ToolStartDecision{Block: true, Code: "TOOL_USE_NOT_IN_POLICY_BUNDLE", Reason: "blocked in test"}
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	running.EmitStdout([]byte("START"))
	select {
	case <-running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("expected provider kill after tool start block")
	}
	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultBlocked {
		t.Fatalf("result: %+v", res)
	}
	if res.Error != "TOOL_USE_NOT_IN_POLICY_BUNDLE: blocked in test" {
		t.Fatalf("block error: %q", res.Error)
	}
	events := drainEvents(t, sess, time.Second)
	var sawWarning bool
	for _, ev := range events {
		if ev.Kind == agentbridge.EventWarning && ev.Text == "tool use blocked by policy" {
			sawWarning = true
		}
	}
	if !sawWarning {
		t.Fatalf("missing policy warning in events: %+v", events)
	}
}

func TestSessionRemovesTempFilesAfterProcessExit(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

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
		TaskID:    "task-tempfile-exit",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
		TempFiles: []string{tempFile.Name(), tempFile.Name()},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	go func() {
		sess.runningForTest().EmitStdout([]byte("DONE"))
		sess.runningForTest().EmitExit(0, nil)
	}()

	res := waitResult(t, sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	if _, err := os.Stat(tempFile.Name()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed after process exit, stat err=%v", err)
	}
	_ = drainEvents(t, sess, time.Second)
}

func TestSessionRemovesTempFilesAfterCancellation(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID:    "task-tempfile-cancel",
		RuntimeID: "rt-1",
		Adapter:   adapter,
		Process:   fake,
		Spawn:     process.Command{Executable: "fake"},
		TempFiles: []string{tempFile.Name()},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	sess.Cancel(nil)
	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("result: %+v", res)
	}
	if _, err := os.Stat(tempFile.Name()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed after cancellation, stat err=%v", err)
	}
	_ = drainEvents(t, sess, time.Second)
}

func TestSessionRemovesTempFilesAfterTimeout(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID:      "task-tempfile-timeout",
		RuntimeID:   "rt-1",
		Adapter:     adapter,
		Process:     fake,
		Spawn:       process.Command{Executable: "fake"},
		HardTimeout: 20 * time.Millisecond,
		TempFiles:   []string{tempFile.Name()},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("result: %+v", res)
	}
	if _, err := os.Stat(tempFile.Name()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed after timeout, stat err=%v", err)
	}
	_ = drainEvents(t, sess, time.Second)
}

// Hard timeout fires when no Result arrives in time.
func TestSessionHardTimeout(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			return nil, nil, nil // produce no events ever
		},
	}
	sess, err := Start(context.Background(), Config{
		TaskID:      "task-2",
		RuntimeID:   "rt-1",
		Adapter:     adapter,
		Process:     fake,
		Spawn:       process.Command{Executable: "fake"},
		HardTimeout: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("status: %s", res.Status)
	}

	// Process must have been killed.
	select {
	case <-sess.runningForTest().KillRecv():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected kill on timeout")
	}
}

// External Cancel beats a later Result event.
func TestSessionCancellation(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(_ agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID: "task-3", RuntimeID: "rt-1", Adapter: adapter,
		Process: fake, Spawn: process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	sess.Cancel(nil)
	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("status: %s", res.Status)
	}
}

// Process exits non-zero without a provider result → failed.
func TestSessionProcessExitNonZero(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(_ agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID: "task-4", RuntimeID: "rt-1", Adapter: adapter,
		Process: fake, Spawn: process.Command{Executable: "fake"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	go sess.runningForTest().EmitExit(137, nil)

	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultFailed {
		t.Fatalf("status: %s", res.Status)
	}
}

// SessionID propagation — first SessionIdentified flows into Result.SessionID.
func TestSessionPropagatesSessionID(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type == "chunk" {
				return []agentbridge.Event{
					{Kind: agentbridge.EventSessionIdentified, SessionID: "sess-abc"},
					{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}},
				}, nil, nil
			}
			return nil, nil, nil
		},
	}
	sess, _ := Start(context.Background(), Config{
		TaskID: "task-5", RuntimeID: "rt-1", Adapter: adapter,
		Process: fake, Spawn: process.Command{Executable: "fake"},
	})
	go sess.runningForTest().EmitStdout([]byte("x"))
	res := waitResult(t, sess, time.Second)
	if res.SessionID != "sess-abc" {
		t.Fatalf("session id: %q", res.SessionID)
	}
}

// Result is emitted exactly once on the Result channel.
func TestSessionResultExactlyOnce(t *testing.T) {
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if raw.Type == "chunk" {
				return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted}}}, nil, nil
			}
			return nil, nil, nil
		},
	}
	sess, _ := Start(context.Background(), Config{
		TaskID: "task-6", RuntimeID: "rt-1", Adapter: adapter,
		Process: fake, Spawn: process.Command{Executable: "fake"},
	})

	go func() {
		sess.runningForTest().EmitStdout([]byte("x"))
		sess.runningForTest().EmitStdout([]byte("y"))
		sess.runningForTest().EmitExit(0, nil)
	}()

	_ = waitResult(t, sess, time.Second)
	// Second read MUST observe a closed channel (no second result).
	select {
	case _, ok := <-sess.Result():
		if ok {
			t.Fatal("Result channel must close after one delivery")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Result channel didn't close")
	}
}
