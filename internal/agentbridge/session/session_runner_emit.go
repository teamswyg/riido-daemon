package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (r *sessionRunner) interruptDelivery(ctx context.Context, ev agentbridge.Event) {
	if note := r.terminateProtocol(ctx); note != "" {
		ev.Err += "; " + note
	}
	if ev.At.IsZero() {
		ev.At = r.cfg.Now()
	}
	r.state, _ = agentbridge.Reduce(r.state, ev, nil)
	r.deliveryInterrupted = true
	_ = killProcess(ctx, r.proc, r.cfg.ProcessKillTimeout)
	r.mirrorEvent(ev)
}

func (r *sessionRunner) mirrorEvent(ev agentbridge.Event) bool {
	select {
	case r.sess.events <- ev:
		return true
	default:
		return false
	}
}

func (r *sessionRunner) emit(ev agentbridge.Event) {
	if r.deliveryInterrupted {
		return
	}
	if ev.At.IsZero() {
		ev.At = r.cfg.Now()
	}
	if r.state.Terminal {
		r.mirrorEvent(ev)
		return
	}
	if ev.Kind == agentbridge.EventCancellation || ev.Kind == agentbridge.EventTimeout {
		r.interruptDelivery(r.ctx, ev)
		return
	}
	if r.ctx.Err() != nil {
		r.interruptDelivery(r.ctx, contextTerminationEvent(r.ctx))
		return
	}
	if ev.Kind == agentbridge.EventLog && ev.LocalOnly {
		if !r.mirrorEvent(ev) {
			r.localDiagnosticsOmitted++
		}
		return
	}
	if ev.Kind.IsSemanticActivity() {
		r.resetIdle()
	}
	select {
	case r.sess.events <- ev:
	case <-r.ctx.Done():
		r.interruptDelivery(r.ctx, contextTerminationEvent(r.ctx))
		return
	case req := <-r.sess.cancel:
		r.interruptDelivery(req.ctx, agentbridge.Event{Kind: agentbridge.EventCancellation, Err: cancellationMessage(req.cause)})
		return
	case <-r.hardC:
		r.interruptDelivery(r.ctx, agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "hard timeout"})
		return
	case <-r.idleC:
		r.onSemanticIdle()
	}
}

func cancellationMessage(cause error) string {
	if cause != nil {
		return cause.Error()
	}
	return "cancelled"
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
