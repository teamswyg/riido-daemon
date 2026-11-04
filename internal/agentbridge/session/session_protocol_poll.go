package session

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (r *sessionRunner) pollProtocol() {
	driver, ok := r.cfg.ProtocolDriver.(agentbridge.ProtocolPoller)
	if !ok {
		return
	}
	events, commands, err := driver.OnPoll(r.ctx, r.io)
	if err != nil {
		events = append(events, agentbridge.Event{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultFailed, Error: err.Error()}})
	}
	r.applyEvents(events, commands)
}

func (r *sessionRunner) terminateProtocol(parent context.Context) string {
	driver, ok := r.cfg.ProtocolDriver.(agentbridge.ProtocolTerminationDriver)
	if !ok {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 3*time.Second)
	defer cancel()
	wait, err := driver.OnTerminate(ctx, r.io)
	if err != nil {
		return err.Error()
	}
	if !wait {
		return ""
	}
	// No event delivery while cancelling: a blocked subscriber must not prevent
	// removing its queued request. Protocol state remains owned by this actor.
	for {
		select {
		case <-ctx.Done():
			return "external request cleanup was not acknowledged; check the provider app"
		case chunk, ok := <-r.stdoutCh:
			if !ok {
				r.stdoutCh = nil
				return "provider closed before queue cleanup acknowledgement"
			}
			raws, err := r.parser.FeedStdout(chunk)
			if err != nil {
				return err.Error()
			}
			for _, raw := range raws {
				events, _, err := r.cfg.ProtocolDriver.OnRaw(ctx, raw, r.io)
				if err != nil {
					return err.Error()
				}
				for _, event := range events {
					if event.Kind == agentbridge.EventResult {
						return event.Result.Error
					}
				}
			}
		case _, ok := <-r.stderrCh:
			if !ok {
				r.stderrCh = nil
			}
		}
	}
}
