package session

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (r *sessionRunner) loop() {
	var poll <-chan time.Time
	if driver, ok := r.cfg.ProtocolDriver.(agentbridge.ProtocolPoller); ok && driver.PollInterval() > 0 {
		ticker := time.NewTicker(driver.PollInterval())
		defer ticker.Stop()
		poll = ticker.C
	}
	for !r.state.Terminal {
		if r.flushDeferredExitWhenDrained() {
			break
		}
		select {
		case <-poll:
			r.pollProtocol()
		case <-r.ctx.Done():
			r.emitAndTerminate(contextTerminationEvent(r.ctx))
		case req := <-r.sess.cancel:
			r.cancel(req)
		case req := <-r.sess.terminal:
			r.terminal(req)
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
