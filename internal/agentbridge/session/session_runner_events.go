package session

import (
	"slices"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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

func (r *sessionRunner) applyEvents(events []agentbridge.Event, cmds []agentbridge.Command) {
	if slices.ContainsFunc(events, r.applyEvent) {
		return
	}
	for _, cmdEvent := range executeCommands(r.proc, r.cfg.Adapter, cmds, r.cfg.ProcessKillTimeout) {
		r.emit(cmdEvent)
	}
}

func (r *sessionRunner) applyEvent(ev agentbridge.Event) bool {
	expanded := []agentbridge.Event{ev}
	if ev.Kind == agentbridge.EventTextDelta {
		expanded = append(expanded, r.telemetry.Feed(ev.Text)...)
	}
	for _, expandedEvent := range expanded {
		r.emit(expandedEvent)
		if r.blockStartedToolIfNeeded(expandedEvent) {
			return true
		}
		var cmds []agentbridge.Command
		r.state, cmds = agentbridge.Reduce(r.state, expandedEvent, r.cfg.AutoApprove)
		for _, cmdEvent := range executeCommands(r.proc, r.cfg.Adapter, cmds, r.cfg.ProcessKillTimeout) {
			r.emit(cmdEvent)
		}
	}
	return false
}

func (r *sessionRunner) blockStartedToolIfNeeded(ev agentbridge.Event) bool {
	if ev.Kind != agentbridge.EventToolCallStarted {
		return false
	}
	decision := decideStartedTool(r.cfg.ToolStartGate, ev.Tool)
	if !decision.Block {
		return false
	}
	blockReason := toolBlockReason(decision)
	for _, cmdEvent := range executeCommands(r.proc, r.cfg.Adapter, []agentbridge.Command{
		{Kind: agentbridge.CommandCancelProvider, Reason: blockReason},
	}, r.cfg.ProcessKillTimeout) {
		r.emit(cmdEvent)
	}
	r.emit(agentbridge.Event{Kind: agentbridge.EventWarning, Text: "tool use blocked by policy", Err: blockReason})
	blocked := agentbridge.Event{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: agentbridge.ResultBlocked,
			Error:  blockReason,
		},
	}
	r.emit(blocked)
	var cmds []agentbridge.Command
	r.state, cmds = agentbridge.Reduce(r.state, blocked, r.cfg.AutoApprove)
	for _, cmdEvent := range executeCommands(r.proc, r.cfg.Adapter, cmds, r.cfg.ProcessKillTimeout) {
		r.emit(cmdEvent)
	}
	return true
}

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

func (r *sessionRunner) emitAndTerminate(synthetic agentbridge.Event) {
	r.applyEvents([]agentbridge.Event{synthetic}, nil)
}
