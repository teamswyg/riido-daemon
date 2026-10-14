package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (r *sessionRunner) cancel(req cancelRequest) {
	reason := "cancelled"
	if req.cause != nil {
		reason = req.cause.Error()
	}
	r.emitAndTerminateWithContext(req.ctx, agentbridge.Event{Kind: agentbridge.EventCancellation, Err: reason})
}
