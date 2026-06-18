package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (r *sessionRunner) emit(ev agentbridge.Event) {
	if ev.At.IsZero() {
		ev.At = r.cfg.Now()
	}
	select {
	case r.sess.events <- ev:
	case <-r.ctx.Done():
	}
	if ev.Kind.IsSemanticActivity() {
		r.resetIdle()
	}
}

func (r *sessionRunner) resetIdle() {
	if r.idleTimer == nil {
		return
	}
	if !r.idleTimer.Stop() {
		select {
		case <-r.idleTimer.C:
		default:
		}
	}
	r.idleTimer.Reset(r.cfg.SemanticIdle)
}
