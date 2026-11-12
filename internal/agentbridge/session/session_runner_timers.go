package session

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (r *sessionRunner) applySemanticActivity(ev agentbridge.Event) bool {
	select {
	case <-r.ctx.Done():
		r.interruptDelivery(r.ctx, contextTerminationEvent(r.ctx))
	case req := <-r.sess.cancel:
		r.cancel(req)
	case <-r.hardC:
		r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "hard timeout"})
	case <-r.idleC:
		r.onSemanticIdle()
	default:
		if ev.At.IsZero() {
			ev.At = r.cfg.Now()
		}
		r.state, _ = agentbridge.Reduce(r.state, ev, nil)
		r.resetIdle()
		return false
	}
	return true
}

func (r *sessionRunner) startTimers() {
	if r.cfg.HardTimeout > 0 {
		r.hardTimer = time.NewTimer(r.cfg.HardTimeout)
		r.hardC = r.hardTimer.C
	}
	if r.cfg.SemanticIdle > 0 {
		r.idleTimer = time.NewTimer(r.cfg.SemanticIdle)
		r.idleC = r.idleTimer.C
	}
}

func (r *sessionRunner) stopTimers() {
	if r.hardTimer != nil {
		r.hardTimer.Stop()
	}
	if r.idleTimer != nil {
		r.idleTimer.Stop()
	}
}
