package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

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
