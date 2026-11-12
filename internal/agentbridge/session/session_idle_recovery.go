package session

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const idleRecoveryTimeout = 2 * time.Second

func (r *sessionRunner) onSemanticIdle() {
	if r.state.Terminal {
		return
	}
	if driver, ok := r.cfg.ProtocolDriver.(agentbridge.ProtocolIdleRecovery); ok && r.state.SessionID != "" {
		r.recoverIdleResult(driver)
	}
	if r.state.Terminal {
		return
	}
	if r.ctx.Err() != nil {
		r.interruptDelivery(r.ctx, contextTerminationEvent(r.ctx))
		return
	}
	r.interruptDelivery(r.ctx, agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "semantic idle timeout"})
}

func (r *sessionRunner) recoverIdleResult(driver agentbridge.ProtocolIdleRecovery) {
	ctx, cancel := context.WithTimeout(r.ctx, idleRecoveryTimeout)
	defer cancel()
	started, err := driver.BeginIdleRecovery(ctx, r.io)
	if err != nil || !started {
		return
	}
	// Drain the provider stream without publishing diagnostics or dispatching
	// tool requests while waiting for the read-only recovery reply.
	for r.stdoutCh != nil {
		select {
		case <-ctx.Done():
			return
		case req := <-r.sess.cancel:
			r.cancel(req)
			return
		case <-r.hardC:
			r.interruptDelivery(r.ctx, agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "hard timeout"})
			return
		case _, ok := <-r.stderrCh:
			if !ok {
				r.stderrCh = nil
			}
		case chunk, ok := <-r.stdoutCh:
			if !ok {
				r.stdoutCh = nil
				return
			}
			raws, err := r.parser.FeedStdout(chunk)
			if err != nil {
				return
			}
			for _, raw := range raws {
				result, done := driver.OnIdleRecoveryRaw(raw)
				if done {
					if result.Status == agentbridge.ResultCompleted && result.SessionID == r.state.SessionID && ctx.Err() == nil {
						r.commitResult(agentbridge.Event{Kind: agentbridge.EventResult, Result: result})
					}
					return
				}
			}
		}
	}
}
