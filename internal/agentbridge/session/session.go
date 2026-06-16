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
	// DefaultProcessKillTimeout bounds provider process kill calls made during
	// session teardown and reducer-driven cancellation.
	DefaultProcessKillTimeout = 5 * time.Second
)

// Config is the input to Start.
type Config struct {
	TaskID    string
	RuntimeID string

	Adapter agentbridge.Adapter
	Process process.Process
	Spawn   process.Command
	Request agentbridge.StartRequest

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

	// ProcessKillTimeout bounds RunningProcess.Kill calls. Zero defaults to
	// DefaultProcessKillTimeout.
	ProcessKillTimeout time.Duration

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
	if cfg.Process == nil {
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
	if cfg.ProcessKillTimeout <= 0 {
		cfg.ProcessKillTimeout = DefaultProcessKillTimeout
	}

	running, err := cfg.Process.Start(ctx, cfg.Spawn)
	if err != nil {
		return nil, fmt.Errorf("session: process start: %w", err)
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
