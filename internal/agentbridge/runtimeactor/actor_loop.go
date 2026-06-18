package runtimeactor

import (
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// run is the actor loop. SOLE owner of in-flight map and per-task state.
func (a *Actor) run(caps []Capability, detectedAt map[string]time.Time) {
	adapters := indexAdapters(a.cfg.Adapters)
	inFlight := map[string]*runningTask{}

	completeCh := make(chan string, 32)

	for {
		select {
		case env := <-a.mailbox:
			switch {
			case env.submit != nil:
				h, err := a.handleSubmit(adapters, caps, detectedAt, inFlight, completeCh, env.submit)
				env.submit.reply <- submitReply{handle: h, err: err}
			case env.cancel != nil:
				env.cancel.reply <- a.handleCancel(inFlight, env.cancel)
			}

		case taskID := <-completeCh:
			delete(inFlight, taskID)

		case msg := <-a.statusCh:
			a.refreshDueCapabilities(msg.ctx, adapters, caps, detectedAt)
			msg.reply <- statusReply{
				status: a.buildStatus(caps, inFlight),
				hb:     a.buildHeartbeat(inFlight),
			}

		case level := <-a.stopReqCh:
			a.drainAndShutdown(lifecycle.NormalizeShutdownLevel(level), inFlight, completeCh)
			close(a.stoppedCh)
			return
		}
	}
}
