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
	for _, cmdEvent := range r.executeCommands(ctx, cmds) {
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
	var accepted bool
	ev, accepted = r.guardSessionIdentity(ev)
	if !accepted {
		return false
	}
	if ev.Kind == agentbridge.EventResult {
		r.commitResult(ev)
		return r.deliveryInterrupted || r.sessionIdentityBlocked
	}
	if ev.Kind == agentbridge.EventSemanticActivity {
		return r.applySemanticActivity(ev)
	}
	approver := r.cfg.AutoApprove
	if ev.Kind == agentbridge.EventToolApprovalNeeded && ev.Tool.RequireHuman {
		ev.AutomaticallyApproved = false
		approver = nil
	} else if ev.Kind == agentbridge.EventToolApprovalNeeded && approver != nil {
		approved := approver(ev.Tool)
		ev.AutomaticallyApproved = approved
		approver = func(agentbridge.ToolRef) bool { return approved }
	}
	if ev.Kind == agentbridge.EventCancellation || ev.Kind == agentbridge.EventTimeout || r.ctx.Err() != nil {
		r.emit(ev)
		return true
	}
	if ev.At.IsZero() {
		ev.At = r.cfg.Now()
	}
	if ev.Kind == agentbridge.EventToolCallStarted {
		// The tool-start gate owns rejection and its notification order.
		r.emit(ev)
		if r.deliveryInterrupted || r.state.Terminal || r.blockStartedToolIfNeeded(ev) {
			return true
		}
	}
	var cmds []agentbridge.Command
	r.state, cmds = agentbridge.Reduce(r.state, ev, approver)
	if ev.Kind != agentbridge.EventToolCallStarted {
		r.emit(ev)
		if r.deliveryInterrupted || r.state.Terminal {
			return r.deliveryInterrupted
		}
	}
	if r.blockPendingApprovalIfNeeded(ev, cmds) {
		return true
	}
	for _, cmdEvent := range r.executeCommands(r.ctx, cmds) {
		r.emit(cmdEvent)
	}
	return r.sessionIdentityBlocked
}

func (r *sessionRunner) emitAndTerminate(synthetic agentbridge.Event) {
	r.emitAndTerminateWithContext(r.ctx, synthetic)
}

func (r *sessionRunner) emitAndTerminateWithContext(ctx context.Context, synthetic agentbridge.Event) {
	r.applyEventsWithContext(ctx, []agentbridge.Event{synthetic}, nil)
}
