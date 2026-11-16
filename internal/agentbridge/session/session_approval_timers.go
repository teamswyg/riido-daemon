package session

import "time"

// Human response time does not consume the provider's execution budget.
// An already exhausted budget remains authoritative before entering the wait.
func (r *sessionRunner) pauseApprovalHardTimeout() func() {
	if r.hardTimer == nil {
		return func() {}
	}
	remaining := time.Until(r.hardDeadline)
	if r.cfg.runDeadline != nil {
		remaining = time.Until(r.cfg.runDeadlineAt)
	}
	if remaining <= 0 {
		return func() {}
	}
	if r.cfg.runDeadline != nil && !r.cfg.runDeadline.Stop() {
		return func() {}
	}
	r.hardTimer.Stop()
	r.hardC = nil
	paused := true
	return func() {
		if !paused {
			return
		}
		paused = false
		if !r.state.Terminal {
			r.hardDeadline = time.Now().Add(remaining)
			r.cfg.runDeadlineAt = r.hardDeadline
			if r.cfg.runDeadline != nil {
				r.cfg.runDeadline.Reset(remaining)
			}
			r.hardTimer.Reset(remaining)
			r.hardC = r.hardTimer.C
		}
	}
}
