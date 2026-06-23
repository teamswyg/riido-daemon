package session

import (
	"context"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (r *sessionRunner) applyEvents(events []agentbridge.Event, cmds []agentbridge.Command) {
	r.applyEventsWithContext(r.ctx, events, cmds)
}

func (r *sessionRunner) applyEventsWithContext(ctx context.Context, events []agentbridge.Event, cmds []agentbridge.Command) {
	if slices.ContainsFunc(events, r.applyEvent) {
		return
	}
	for _, cmdEvent := range executeCommands(ctx, r.proc, r.cfg.Adapter, cmds, r.cfg.ProcessKillTimeout) {
		r.emit(cmdEvent)
	}
}

func (r *sessionRunner) applyEvent(ev agentbridge.Event) bool {
	expanded := []agentbridge.Event{ev}
	if ev.Kind == agentbridge.EventTextDelta {
		visible, telemetryEvents := r.telemetry.FilterTextDelta(ev.Text)
		expanded = telemetryEvents
		if visible != "" {
			ev.Text = visible
			expanded = append([]agentbridge.Event{ev}, expanded...)
		}
	}
	return slices.ContainsFunc(expanded, r.applyExpandedEvent)
}

func (r *sessionRunner) applyExpandedEvent(ev agentbridge.Event) bool {
	r.emit(ev)
	if r.blockStartedToolIfNeeded(ev) {
		return true
	}
	var cmds []agentbridge.Command
	r.state, cmds = agentbridge.Reduce(r.state, ev, r.cfg.AutoApprove)
	if r.blockPendingApprovalIfNeeded(ev, cmds) {
		return true
	}
	for _, cmdEvent := range executeCommands(r.ctx, r.proc, r.cfg.Adapter, cmds, r.cfg.ProcessKillTimeout) {
		r.emit(cmdEvent)
	}
	return false
}

func (r *sessionRunner) emitAndTerminate(synthetic agentbridge.Event) {
	r.emitAndTerminateWithContext(r.ctx, synthetic)
}

func (r *sessionRunner) emitAndTerminateWithContext(ctx context.Context, synthetic agentbridge.Event) {
	r.applyEventsWithContext(ctx, []agentbridge.Event{synthetic}, nil)
}
