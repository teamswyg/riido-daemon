package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (r *sessionRunner) feed(raws []agentbridge.RawEvent) {
	for _, raw := range raws {
		events, cmds, err := r.translateRaw(raw)
		if err != nil {
			r.emit(agentbridge.Event{Kind: agentbridge.EventError, Err: err.Error()})
			continue
		}
		r.applyEvents(events, cmds)
		if r.state.Terminal {
			return
		}
	}
}

func (r *sessionRunner) translateRaw(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if r.cfg.ProtocolDriver != nil {
		return r.cfg.ProtocolDriver.OnRaw(r.ctx, raw, r.io)
	}
	return r.cfg.Adapter.Translate(raw)
}
