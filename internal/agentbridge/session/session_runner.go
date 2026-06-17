package session

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type sessionRunner struct {
	ctx  context.Context
	cfg  Config
	sess *Session
	proc process.RunningProcess

	parser    agentbridge.Parser
	state     agentbridge.State
	telemetry *agentbridge.TelemetryParser
	io        *protocolIOImpl

	startedAt time.Time
	stdoutCh  <-chan []byte
	stderrCh  <-chan []byte
	exitedCh  <-chan process.ExitStatus

	hardTimer *time.Timer
	hardC     <-chan time.Time
	idleTimer *time.Timer
	idleC     <-chan time.Time

	deferredExit *process.ExitStatus
}

func run(ctx context.Context, cfg Config, sess *Session, proc process.RunningProcess) {
	defer close(sess.done)
	defer close(sess.events)
	defer close(sess.result)

	runner := newSessionRunner(ctx, cfg, sess, proc)
	defer runner.stopTimers()
	runner.startProtocol()
	defer runner.closeProtocol()

	runner.loop()
	runner.finish()
}

func newSessionRunner(ctx context.Context, cfg Config, sess *Session, proc process.RunningProcess) *sessionRunner {
	state := agentbridge.NewState()
	state.LastSemanticActivity = cfg.Now()
	startedAt := cfg.Now()

	runner := &sessionRunner{
		ctx:       ctx,
		cfg:       cfg,
		sess:      sess,
		proc:      proc,
		parser:    cfg.Adapter.NewParser(),
		state:     state,
		telemetry: agentbridge.NewTelemetryParser(),
		io:        newProtocolIO(proc),
		startedAt: startedAt,
		stdoutCh:  proc.Stdout(),
		stderrCh:  proc.Stderr(),
		exitedCh:  proc.Exited(),
	}
	if cfg.HardTimeout > 0 {
		runner.hardTimer = time.NewTimer(cfg.HardTimeout)
		runner.hardC = runner.hardTimer.C
	}
	if cfg.SemanticIdle > 0 {
		runner.idleTimer = time.NewTimer(cfg.SemanticIdle)
		runner.idleC = runner.idleTimer.C
	}
	return runner
}

func (r *sessionRunner) stopTimers() {
	if r.hardTimer != nil {
		r.hardTimer.Stop()
	}
	if r.idleTimer != nil {
		r.idleTimer.Stop()
	}
}

func (r *sessionRunner) loop() {
	for !r.state.Terminal {
		if r.flushDeferredExitWhenDrained() {
			break
		}
		select {
		case <-r.ctx.Done():
			r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventCancellation, Err: r.ctx.Err().Error()})
		case req := <-r.sess.cancel:
			r.cancel(req)
		case <-r.hardC:
			r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "hard timeout"})
		case <-r.idleC:
			r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "semantic idle timeout"})
		case chunk, ok := <-r.stdoutCh:
			r.consumeStdout(chunk, ok)
		case chunk, ok := <-r.stderrCh:
			r.consumeStderr(chunk, ok)
		case status, ok := <-r.exitedCh:
			r.deferExit(status, ok)
		}
	}
}

func (r *sessionRunner) finish() {
	_ = killProcess(r.ctx, r.proc, r.cfg.ProcessKillTimeout)
	drain(r.stdoutCh)
	drain(r.stderrCh)
	for _, ev := range cleanupTempFiles(r.cfg.TempFiles) {
		r.emit(ev)
	}
	r.sess.result <- r.finalResult()
}
