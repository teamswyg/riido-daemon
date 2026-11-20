package session

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const idleRecoveryTimeout = 2 * time.Second

func (r *sessionRunner) onProviderIdle() {
	if driver, ok := r.cfg.ProtocolDriver.(agentbridge.ProtocolIdleGrace); ok && driver.ConsumeIdleGrace() {
		r.resetIdle()
		return
	}
	r.onSemanticIdle()
}

func (r *sessionRunner) onSemanticIdle() {
	if r.state.Terminal || r.idleRecoveryTimer != nil {
		return
	}
	if driver, ok := r.cfg.ProtocolDriver.(agentbridge.ProtocolIdleRecovery); ok && r.state.SessionID != "" {
		started, err := driver.BeginIdleRecovery(r.ctx, r.io)
		if err == nil && started {
			// Keep the ordinary actor loop alive: progress, approvals, cancellation,
			// process exit and terminal notifications must not be lost during reads.
			r.idleRecoveryTimer = time.NewTimer(idleRecoveryTimeout)
			r.idleRecoveryC = r.idleRecoveryTimer.C
			return
		}
	}
	r.finishIdleTimeout()
}

func (r *sessionRunner) consumeIdleRecovery(raw agentbridge.RawEvent) bool {
	if r.idleRecoveryTimer == nil {
		return false
	}
	if r.ctx.Err() != nil {
		r.interruptDelivery(r.ctx, contextTerminationEvent(r.ctx))
		return true
	}
	select {
	case req := <-r.sess.cancel:
		r.cancel(req)
		return true
	default:
	}
	select {
	case <-r.hardC:
		r.interruptDelivery(r.ctx, agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "hard timeout"})
		return true
	default:
	}
	select {
	case <-r.idleRecoveryC:
		r.finishIdleTimeout()
		return true
	default:
	}
	driver := r.cfg.ProtocolDriver.(agentbridge.ProtocolIdleRecovery)
	result, done := driver.OnIdleRecoveryRaw(raw)
	if !done {
		return false
	}
	r.stopIdleRecovery()
	if result.Status == agentbridge.ResultCompleted && result.SessionID == r.state.SessionID {
		r.commitResult(agentbridge.Event{Kind: agentbridge.EventResult, Result: result})
		return true
	}
	if grace, ok := driver.(agentbridge.ProtocolIdleGrace); ok && grace.ConsumeIdleGrace() {
		r.resetIdle()
		r.mirrorEvent(agentbridge.Event{Kind: agentbridge.EventLog, LocalOnly: true, Text: "provider idle recovery: current turn still running; bounded grace granted"})
		return true
	}
	r.finishIdleTimeout()
	return true
}

func (r *sessionRunner) stopIdleRecovery() {
	if r.idleRecoveryTimer == nil {
		return
	}
	r.idleRecoveryTimer.Stop()
	r.idleRecoveryTimer = nil
	r.idleRecoveryC = nil
	r.cfg.ProtocolDriver.(agentbridge.ProtocolIdleRecovery).EndIdleRecovery()
}

func (r *sessionRunner) finishIdleTimeout() {
	r.stopIdleRecovery()
	if r.state.Terminal {
		return
	}
	if r.ctx.Err() != nil {
		r.interruptDelivery(r.ctx, contextTerminationEvent(r.ctx))
		return
	}
	r.interruptDelivery(r.ctx, agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "semantic idle timeout"})
}
