package supervisor

import "time"

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func resetTimer(t *time.Timer, d time.Duration) {
	stopTimer(t)
	t.Reset(d)
}

func (a *Actor) nextPollInterval(claimed bool, inFlightCount int) time.Duration {
	if claimed || inFlightCount > 0 {
		return a.cfg.PollEvery
	}
	return a.cfg.IdlePollEvery
}
