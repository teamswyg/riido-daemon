// Package session implements the run-scope session actor: one goroutine
// that owns one provider session's State. It wires Process → Parser →
// Adapter.Translate → Reducer → emit Events/Result.
//
// The actor is the single owner of the agentbridge.State for its run.
// The reducer is called inline; no other goroutine ever touches State.
// No sync.Mutex / sync.RWMutex is used: backpressure and shutdown are
// expressed via bounded channels per docs/20-domain/provider-runtime.md §7.5.
package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

const (
	// DefaultEventBuffer is the C4 Provider Runtime backpressure buffer
	// recorded in docs/20-domain/provider-runtime.md §7.5.
	DefaultEventBuffer = 256
	// DefaultResultBuffer stores the single terminal result for a session.
	DefaultResultBuffer = 1
)

// Config is the input to Start.
type Config struct {
	TaskID    string
	RuntimeID string

	Adapter agentbridge.Adapter
	Process process.Process
	Spawn   process.Command
	// Running reuses an already-started provider process. When set, Process is
	// not used to spawn a new process.
	Running process.RunningProcess
	// KeepProcessAlive keeps Running alive after provider terminal results so a
	// persistent runtime can submit another thread/turn on the same transport.
	// Cancellation, timeout, blocked runs, and process exits still kill/reap.
	KeepProcessAlive bool
	Request          agentbridge.StartRequest

	// HardTimeout is the wall-clock deadline for the entire run. Zero disables.
	HardTimeout time.Duration
	// SemanticIdle is the idle watchdog timeout — no semantic event for this
	// duration triggers a timeout. Zero disables.
	SemanticIdle time.Duration

	// AutoApprove decides tool approval (see agentbridge.AutoApprover).
	// Default (nil) requires explicit human approval.
	AutoApprove agentbridge.AutoApprover
	// ToolStartGate decides whether an already-started provider tool call must
	// fail closed because the provider did not expose an approval round-trip.
	ToolStartGate agentbridge.ToolStartGate

	// EventBuffer / ResultBuffer override default channel capacities.
	EventBuffer  int
	ResultBuffer int

	// ProtocolDriver is an optional hook that takes over RawEvent
	// routing when the provider transport needs an active handshake
	// (e.g. Codex JSON-RPC initialize/initialized/thread/start/turn).
	// When non-nil it REPLACES Adapter.Translate as the raw → events
	// converter — see protocol_driver.go for the lifecycle contract.
	// When nil the session uses Adapter.Translate (legacy path; all
	// existing Claude / OpenClaw / Cursor / fake adapters).
	ProtocolDriver ProtocolDriver

	// TempFiles are adapter-owned files that exist only for this provider
	// process. The session actor removes them after the process exits or the
	// run is cancelled/timed out, so provider secrets/config do not linger in
	// the task workdir or system temp directory.
	TempFiles []string

	// now is injected for deterministic tests; defaults to time.Now.
	Now func() time.Time
}

// Session is the public handle the caller uses to consume events and the
// terminal result.
type Session struct {
	events chan agentbridge.Event
	result chan agentbridge.Result
	cancel chan error
	done   chan struct{}

	running process.RunningProcess
}

// Events returns the event stream. It is closed when the session terminates.
func (s *Session) Events() <-chan agentbridge.Event { return s.events }

// Result returns a single-value channel that delivers the terminal Result
// exactly once and is then closed.
func (s *Session) Result() <-chan agentbridge.Result { return s.result }

// Done returns a channel that is closed when the session terminates. Use
// it from supervisors that want to know "this run is over" without
// consuming the terminal Result (which is for the original caller).
func (s *Session) Done() <-chan struct{} { return s.done }

// Cancel signals the session to terminate as ResultCancelled. Cause may
// be nil. Safe to call from any goroutine.
func (s *Session) Cancel(cause error) {
	select {
	case s.cancel <- cause:
	default:
	}
}

// runningForTest exposes the underlying process for whitebox tests in
// the same package. Production code does not use it.
func (s *Session) runningForTest() *process.FakeRunning {
	r, _ := s.running.(*process.FakeRunning)
	return r
}

// Start spawns the session actor and returns its Session handle. The
// caller MUST drain Events() until it is closed; otherwise the actor
// will block on send and the cancellation/result path will stall.
func Start(ctx context.Context, cfg Config) (*Session, error) {
	if cfg.Adapter == nil {
		return nil, errors.New("session: Adapter is required")
	}
	if cfg.Process == nil && cfg.Running == nil {
		return nil, errors.New("session: Process is required")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.EventBuffer <= 0 {
		cfg.EventBuffer = DefaultEventBuffer
	}
	if cfg.ResultBuffer <= 0 {
		cfg.ResultBuffer = DefaultResultBuffer
	}

	running := cfg.Running
	if running == nil {
		var err error
		running, err = cfg.Process.Start(ctx, cfg.Spawn)
		if err != nil {
			return nil, fmt.Errorf("session: process start: %w", err)
		}
	}

	sess := &Session{
		events:  make(chan agentbridge.Event, cfg.EventBuffer),
		result:  make(chan agentbridge.Result, cfg.ResultBuffer),
		cancel:  make(chan error, 1),
		done:    make(chan struct{}),
		running: running,
	}

	go run(ctx, cfg, sess, running)
	return sess, nil
}

func run(ctx context.Context, cfg Config, sess *Session, proc process.RunningProcess) {
	defer close(sess.done)
	defer close(sess.events)
	defer close(sess.result)

	parser := cfg.Adapter.NewParser()
	state := agentbridge.NewState()
	telemetry := agentbridge.NewTelemetryParser()
	state.LastSemanticActivity = cfg.Now()
	startedAt := cfg.Now()

	// ProtocolIO is bound to this run's process. We thread it through
	// every driver hook so the driver never holds a reference across
	// goroutine boundaries.
	io := newProtocolIO(proc)

	stdoutCh := proc.Stdout()
	stderrCh := proc.Stderr()
	exitedCh := proc.Exited()

	var hardTimer *time.Timer
	var hardC <-chan time.Time
	if cfg.HardTimeout > 0 {
		hardTimer = time.NewTimer(cfg.HardTimeout)
		hardC = hardTimer.C
		defer hardTimer.Stop()
	}

	var idleTimer *time.Timer
	var idleC <-chan time.Time
	if cfg.SemanticIdle > 0 {
		idleTimer = time.NewTimer(cfg.SemanticIdle)
		idleC = idleTimer.C
		defer idleTimer.Stop()
	}
	resetIdle := func() {
		if idleTimer != nil {
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(cfg.SemanticIdle)
		}
	}

	emit := func(ev agentbridge.Event) {
		if ev.At.IsZero() {
			ev.At = cfg.Now()
		}
		select {
		case sess.events <- ev:
		case <-ctx.Done():
		}
		// B1: telemetry progress (`<riido_log>` code 1001 "thinking" etc.) is the
		// agent's self-narration, not ground-truth provider activity, so it must
		// NOT keep the semantic-idle watchdog alive — otherwise a run that only
		// emits progress pings never idle-times-out. Real output (text, tool
		// calls, usage, reasoning) still resets it. Scoped to this watchdog only;
		// the global IsSemanticActivity / reducer LastSemanticActivity are unchanged.
		if ev.Kind.IsSemanticActivity() && ev.Kind != agentbridge.EventProgress {
			resetIdle()
		}
	}

	applyEvents := func(events []agentbridge.Event, cmds []agentbridge.Command) {
		for _, ev := range events {
			expanded := []agentbridge.Event{ev}
			if ev.Kind == agentbridge.EventTextDelta {
				// A1: strip `<riido_log>…<end>` telemetry from the forwarded text
				// and surface the parsed progress events instead. A delta that is
				// pure telemetry (nothing left after stripping) is dropped so it
				// neither leaks raw markers nor — being EventTextDelta — resets the
				// idle watchdog (works with the EventProgress exclusion in emit).
				cleaned, telemetryEvents := telemetry.Feed(ev.Text)
				if strings.TrimSpace(cleaned) == "" {
					expanded = telemetryEvents
				} else {
					ev.Text = cleaned
					expanded = append([]agentbridge.Event{ev}, telemetryEvents...)
				}
			}
			for _, expandedEvent := range expanded {
				emit(expandedEvent)
				if expandedEvent.Kind == agentbridge.EventToolCallStarted {
					if decision := decideStartedTool(cfg.ToolStartGate, expandedEvent.Tool); decision.Block {
						blockReason := toolBlockReason(decision)
						for _, cmdEvent := range executeCommands(proc, cfg.Adapter, []agentbridge.Command{{Kind: agentbridge.CommandCancelProvider, Reason: blockReason}}) {
							emit(cmdEvent)
						}
						emit(agentbridge.Event{Kind: agentbridge.EventWarning, Text: "tool use blocked by policy", Err: blockReason})
						blocked := agentbridge.Event{
							Kind: agentbridge.EventResult,
							Result: agentbridge.Result{
								Status: agentbridge.ResultBlocked,
								Error:  blockReason,
							},
						}
						emit(blocked)
						var blockCmds []agentbridge.Command
						state, blockCmds = agentbridge.Reduce(state, blocked, cfg.AutoApprove)
						for _, cmdEvent := range executeCommands(proc, cfg.Adapter, blockCmds) {
							emit(cmdEvent)
						}
						return
					}
				}
				var newCmds []agentbridge.Command
				state, newCmds = agentbridge.Reduce(state, expandedEvent, cfg.AutoApprove)
				for _, cmdEvent := range executeCommands(proc, cfg.Adapter, newCmds) {
					emit(cmdEvent)
				}
			}
		}
		for _, cmdEvent := range executeCommands(proc, cfg.Adapter, cmds) {
			emit(cmdEvent)
		}
	}

	feed := func(raws []agentbridge.RawEvent) {
		for _, raw := range raws {
			var events []agentbridge.Event
			var cmds []agentbridge.Command
			var err error
			if cfg.ProtocolDriver != nil {
				events, cmds, err = cfg.ProtocolDriver.OnRaw(ctx, raw, io)
			} else {
				events, cmds, err = cfg.Adapter.Translate(raw)
			}
			if err != nil {
				emit(agentbridge.Event{Kind: agentbridge.EventError, Err: err.Error()})
				continue
			}
			applyEvents(events, cmds)
			if state.Terminal {
				return
			}
		}
	}

	emitAndTerminate := func(synthetic agentbridge.Event) {
		applyEvents([]agentbridge.Event{synthetic}, nil)
	}

	// OnStart — drive the initial handshake frames before entering the
	// main loop. A driver error here is fatal: emit the error and
	// terminate with ResultFailed without consulting the process.
	if cfg.ProtocolDriver != nil {
		if err := cfg.ProtocolDriver.OnStart(ctx, io); err != nil {
			applyEvents([]agentbridge.Event{
				{Kind: agentbridge.EventError, Err: err.Error()},
				{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultFailed, Error: err.Error()}},
			}, nil)
		}
	}

	// Defer OnClose so it runs regardless of how we exit the loop.
	defer func() {
		if cfg.ProtocolDriver != nil {
			_ = cfg.ProtocolDriver.OnClose(ctx, io)
		}
	}()

	// Once Exited fires we capture the status but keep draining stdout/stderr
	// until those channels close. Only after both streams are fully consumed
	// do we apply the deferred ProcessExit event — otherwise a terminal
	// provider Result sitting in the stdout buffer could be lost to an
	// out-of-order exit.
	var deferredExit *process.ExitStatus

	for !state.Terminal {
		if stdoutCh == nil && stderrCh == nil && deferredExit != nil {
			if raws, err := parser.Close(); err == nil {
				feed(raws)
			}
			if state.Terminal {
				break
			}
			// Let the protocol driver react to process exit BEFORE we
			// apply the canonical reducer policy. Typical use: fail
			// pending RPC requests with Error events so callers don't
			// leak.
			if cfg.ProtocolDriver != nil {
				exitStatus := agentbridge.ProcessExitStatus{Code: deferredExit.Code}
				if deferredExit.Err != nil {
					exitStatus.Err = deferredExit.Err.Error()
				}
				driverEvents, err := cfg.ProtocolDriver.OnProcessExit(ctx, exitStatus, io)
				if err != nil {
					emit(agentbridge.Event{Kind: agentbridge.EventError, Err: err.Error()})
				}
				if len(driverEvents) > 0 {
					applyEvents(driverEvents, nil)
				}
				if state.Terminal {
					break
				}
			}
			errMsg := ""
			if deferredExit.Err != nil {
				errMsg = deferredExit.Err.Error()
			}
			applyEvents([]agentbridge.Event{{Kind: agentbridge.EventProcessExit, ExitCode: deferredExit.Code, Err: errMsg}}, nil)
			if !state.Terminal {
				applyEvents([]agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultAborted, Error: "process exited without provider result"}}}, nil)
			}
			break
		}

		select {
		case <-ctx.Done():
			emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventCancellation, Err: ctx.Err().Error()})

		case cause := <-sess.cancel:
			reason := "cancelled"
			if cause != nil {
				reason = cause.Error()
			}
			emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventCancellation, Err: reason})

		case <-hardC:
			emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: fmt.Sprintf("run hard timeout exceeded (%s)", cfg.HardTimeout)})

		case <-idleC:
			emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: fmt.Sprintf("no provider progress for %s (semantic idle timeout)", cfg.SemanticIdle)})

		case chunk, ok := <-stdoutCh:
			if !ok {
				stdoutCh = nil
				continue
			}
			raws, err := parser.FeedStdout(chunk)
			if err != nil {
				emit(agentbridge.Event{Kind: agentbridge.EventError, Err: err.Error()})
				continue
			}
			feed(raws)

		case chunk, ok := <-stderrCh:
			if !ok {
				stderrCh = nil
				continue
			}
			raws, err := parser.FeedStderr(chunk)
			if err != nil {
				continue
			}
			feed(raws)

		case status, ok := <-exitedCh:
			if !ok {
				exitedCh = nil
				continue
			}
			st := status
			deferredExit = &st
			exitedCh = nil
		}
	}

	keepProcess := cfg.KeepProcessAlive && canKeepProviderProcess(state.Result.Status) && deferredExit == nil
	if !keepProcess {
		// Make sure the process is reaped if we terminated for a non-exit reason.
		_ = proc.Kill(context.Background())
		// Drain remaining stdout/stderr non-blockingly so the fake process doesn't block.
		drain(stdoutCh)
		drain(stderrCh)
	}
	for _, ev := range cleanupTempFiles(cfg.TempFiles) {
		emit(ev)
	}

	finalResult := state.Result
	finalResult.SessionID = state.SessionID
	finalResult.Usage = state.Usage
	if finalResult.Workdir == "" {
		finalResult.Workdir = cfg.Spawn.Dir
	}
	if finalResult.StartedAt.IsZero() {
		finalResult.StartedAt = startedAt
	}
	if finalResult.FinishedAt.IsZero() {
		finalResult.FinishedAt = cfg.Now()
	}
	sess.result <- finalResult
}

func canKeepProviderProcess(status agentbridge.ResultStatus) bool {
	switch status {
	case agentbridge.ResultCompleted, agentbridge.ResultFailed:
		return true
	default:
		return false
	}
}

func cleanupTempFiles(paths []string) []agentbridge.Event {
	var out []agentbridge.Event
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			out = append(out, agentbridge.Event{
				Kind: agentbridge.EventWarning,
				Text: "adapter temp file cleanup failed",
				Err:  fmt.Sprintf("%s: %v", path, err),
			})
		}
	}
	return out
}

func executeCommands(proc process.RunningProcess, adapter agentbridge.Adapter, cmds []agentbridge.Command) []agentbridge.Event {
	var out []agentbridge.Event
	for _, c := range cmds {
		switch c.Kind {
		case agentbridge.CommandCancelProvider:
			_ = proc.Kill(context.Background())
		case agentbridge.CommandWriteProviderInput:
			if len(c.Input) > 0 {
				if err := proc.WriteStdin(c.Input); err != nil {
					out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider input write failed", Err: err.Error()})
				}
			}
		case agentbridge.CommandApproveTool, agentbridge.CommandRejectTool:
			builder, ok := adapter.(agentbridge.ProviderInputBuilder)
			if !ok {
				out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider approval command has no input builder"})
				continue
			}
			input, err := builder.BuildProviderInput(c)
			if err != nil {
				out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider approval command build failed", Err: err.Error()})
				continue
			}
			if len(input) > 0 {
				if err := proc.WriteStdin(input); err != nil {
					out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider approval command write failed", Err: err.Error()})
				}
			}
		case agentbridge.CommandFlushEvents,
			agentbridge.CommandPersistSession,
			agentbridge.CommandStartProvider:
			// Other commands are no-ops at this layer; the supervisor /
			// runtime actor (to be added) will route them. For the session
			// actor we just need to ensure CancelProvider terminates the
			// child process, which Reduce already emits on Cancel/Timeout.
		}
	}
	return out
}

func decideStartedTool(gate agentbridge.ToolStartGate, tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
	if gate == nil {
		return agentbridge.ToolStartDecision{}
	}
	return gate(tool)
}

func toolBlockReason(decision agentbridge.ToolStartDecision) string {
	if decision.Code == "" {
		return decision.Reason
	}
	if decision.Reason == "" {
		return decision.Code
	}
	return decision.Code + ": " + decision.Reason
}

func drain(ch <-chan []byte) {
	if ch == nil {
		return
	}
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
