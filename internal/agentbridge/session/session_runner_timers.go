package session

import "time"

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
