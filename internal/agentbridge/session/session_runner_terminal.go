package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (r *sessionRunner) terminal(req terminalRequest) {
	r.applyEventsWithContext(req.ctx, []agentbridge.Event{{
		Kind:   agentbridge.EventResult,
		Result: req.result,
	}}, nil)
}
