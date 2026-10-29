package runtimeactor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type capabilityProbe struct {
	generation  uint64
	fingerprint string
	cancel      context.CancelFunc
	pending     []*submitMsg
	waiters     []chan error
}

type capabilityProbeResult struct {
	provider    string
	generation  uint64
	fingerprint string
	capability  Capability
	models      []RuntimeModel
	err         error
}

// This state is owned only by run. Probe workers receive immutable inputs.
type capabilityRefresh struct {
	next       uint64
	probes     map[string]*capabilityProbe
	results    chan capabilityProbeResult
	pendingIDs map[string]bool
}

func newCapabilityRefresh(providerCount int) *capabilityRefresh {
	return &capabilityRefresh{probes: map[string]*capabilityProbe{}, results: make(chan capabilityProbeResult, providerCount), pendingIDs: map[string]bool{}}
}

func (r *capabilityRefresh) start(a *Actor, adapter agentbridge.Adapter, capView Capability) *capabilityProbe {
	if probe := r.probes[capView.Provider]; probe != nil {
		return probe
	}
	r.next++
	ctx, cancel := context.WithTimeout(context.Background(), DefaultCapabilityProbeTimeout)
	probe := &capabilityProbe{generation: r.next, fingerprint: capView.CapabilityFingerprint, cancel: cancel}
	r.probes[capView.Provider] = probe
	// Capture time in the actor; injected clocks need not be goroutine safe.
	observedAt := a.cfg.Now()
	go func() {
		defer cancel()
		capability, err := a.detectCapabilityAt(ctx, adapter, observedAt)
		result := capabilityProbeResult{provider: capView.Provider, generation: probe.generation, fingerprint: probe.fingerprint, capability: capability, err: err}
		if err == nil && capability.Available && a.cfg.RefreshModels != nil {
			result.models = a.cfg.RefreshModels(ctx, capability)
		}
		select {
		case r.results <- result:
		case <-a.stoppedCh:
		}
	}()
	return probe
}

func (r *capabilityRefresh) deferSubmit(a *Actor, adapters map[string]agentbridge.Adapter, caps []Capability, detectedAt map[string]time.Time, msg *submitMsg) bool {
	provider := string(msg.req.Provider)
	idx := capabilityIndexForProvider(caps, provider)
	if idx < 0 {
		return false
	}
	capView := caps[idx]
	if r.probes[provider] == nil && !a.capabilityRefreshDue(capView, detectedAt[provider]) {
		return false
	}
	if r.pendingIDs[msg.req.ID] {
		msg.reply <- submitReply{err: ErrDuplicateTaskID}
		return true
	}
	if len(r.pendingIDs) >= a.cfg.MailboxSize {
		msg.reply <- submitReply{err: ErrSlotExhausted}
		return true
	}
	probe := r.start(a, adapters[provider], capView)
	probe.pending = append(probe.pending, msg)
	r.pendingIDs[msg.req.ID] = true
	return true
}

func (r *capabilityRefresh) finish(a *Actor, result capabilityProbeResult, caps []Capability, detectedAt map[string]time.Time, inFlight map[string]*runningTask) []*submitMsg {
	probe := r.probes[result.provider]
	if probe == nil || probe.generation != result.generation || probe.fingerprint != result.fingerprint {
		return nil
	}
	delete(r.probes, result.provider)
	probe.cancel()
	idx := capabilityIndexForProvider(caps, result.provider)
	if idx >= 0 && caps[idx].CapabilityFingerprint == result.fingerprint {
		observed := result.capability
		if result.err != nil {
			observed = capabilityWithProbeFailure(caps[idx])
		} else {
			observed = retainCapabilityProbeIdentity(caps[idx], observed)
		}
		blockRuntimeDriftedTasks(context.Background(), inFlight, observed)
		caps[idx] = observed
		if result.err == nil && len(result.models) > 0 {
			a.cfg.Models = append([]RuntimeModel(nil), result.models...)
		}
		detectedAt[result.provider] = a.cfg.Now()
	} else {
		for _, reply := range probe.waiters {
			reply <- ErrRuntimePinViolated
		}
		for _, msg := range probe.pending {
			delete(r.pendingIDs, msg.req.ID)
			msg.reply <- submitReply{err: ErrRuntimePinViolated}
		}
		return nil
	}
	for _, reply := range probe.waiters {
		reply <- result.err
	}
	for _, msg := range probe.pending {
		delete(r.pendingIDs, msg.req.ID)
	}
	return probe.pending
}

func (r *capabilityRefresh) cancelPending(taskID string) bool {
	for _, probe := range r.probes {
		for idx, msg := range probe.pending {
			if msg.req.ID == taskID {
				probe.pending = append(probe.pending[:idx], probe.pending[idx+1:]...)
				delete(r.pendingIDs, taskID)
				msg.reply <- submitReply{err: context.Canceled}
				return true
			}
		}
	}
	return false
}

func (r *capabilityRefresh) stop() {
	for _, probe := range r.probes {
		probe.cancel()
		for _, reply := range probe.waiters {
			reply <- ErrActorStopped
		}
		for _, msg := range probe.pending {
			msg.reply <- submitReply{err: ErrActorStopped}
		}
	}
}
