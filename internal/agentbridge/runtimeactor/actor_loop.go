package runtimeactor

import (
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// run is the actor loop. SOLE owner of in-flight map and per-task state.
func (a *Actor) run(caps []Capability, detectedAt map[string]time.Time) {
	adapters := indexAdapters(a.cfg.Adapters)
	inFlight := map[string]*runningTask{}

	completeCh := make(chan string, 32)
	refresh := newCapabilityRefresh(len(caps))
	defer refresh.stop()
	var refreshTick <-chan time.Time
	if a.cfg.CapabilityRefreshEvery > 0 {
		ticker := time.NewTicker(a.cfg.CapabilityRefreshEvery)
		defer ticker.Stop()
		refreshTick = ticker.C
	}

	for {
		select {
		case env := <-a.mailbox:
			switch {
			case env.submit != nil:
				if err := a.validateSubmit(inFlight, env.submit); err != nil {
					env.submit.reply <- submitReply{err: preflightError("slot_exhausted", errors.Is(err, ErrSlotExhausted), false, err)}
					continue
				}
				if refresh.deferSubmit(a, adapters, caps, detectedAt, env.submit) {
					continue
				}
				h, err := a.handleSubmit(adapters, caps, detectedAt, inFlight, completeCh, env.submit)
				env.submit.reply <- submitReply{handle: h, err: err}
			case env.refresh != nil:
				idx := capabilityIndexForProvider(caps, env.refresh.provider)
				if idx < 0 {
					env.refresh.reply <- ErrUnknownProvider
				} else {
					probe := refresh.start(a, adapters[env.refresh.provider], caps[idx])
					if len(probe.waiters) >= a.cfg.MailboxSize {
						env.refresh.reply <- ErrSlotExhausted
					} else {
						probe.waiters = append(probe.waiters, env.refresh.reply)
					}
				}
			case env.cancel != nil:
				if refresh.cancelPending(env.cancel.taskID) {
					env.cancel.reply <- nil
				} else {
					env.cancel.reply <- a.handleCancel(inFlight, env.cancel)
				}
			}

		case taskID := <-completeCh:
			delete(inFlight, taskID)

		case msg := <-a.statusCh:
			refresh.scheduleDue(a, adapters, caps, detectedAt)
			msg.reply <- statusReply{
				status: a.buildStatus(caps, inFlight),
				hb:     a.buildHeartbeat(inFlight),
			}

		case result := <-refresh.results:
			pending := refresh.finish(a, result, caps, detectedAt, inFlight)
			for _, msg := range pending {
				h, err := a.handleSubmit(adapters, caps, detectedAt, inFlight, completeCh, msg)
				msg.reply <- submitReply{handle: h, err: err}
			}
		case <-refreshTick:
			refresh.scheduleDue(a, adapters, caps, detectedAt)

		case level := <-a.stopReqCh:
			a.drainAndShutdown(lifecycle.NormalizeShutdownLevel(level), inFlight, completeCh)
			close(a.stoppedCh)
			return
		}
	}
}
