package runtimeactor

import (
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) handleSubmit(
	adapters map[string]agentbridge.Adapter,
	caps []Capability,
	detectedAt map[string]time.Time,
	inFlight map[string]*runningTask,
	completeCh chan<- string,
	msg *submitMsg,
) (*SessionHandle, error) {
	if a.stopping.Load() {
		return nil, ErrActorStopped
	}
	if err := msg.ctx.Err(); err != nil {
		return nil, err
	}
	if err := a.validateSubmit(inFlight, msg); err != nil {
		return nil, preflightError("slot_exhausted", errors.Is(err, ErrSlotExhausted), false, err)
	}
	adapter, capView, err := a.submitCapability(msg, adapters, caps, detectedAt)
	if err != nil {
		return nil, err
	}
	blockRuntimeDriftedTasks(msg.ctx, inFlight, capView)
	if msg.req.ExpectedCapabilityFingerprint != "" && msg.req.ExpectedCapabilityFingerprint != capView.CapabilityFingerprint {
		return nil, preflightError("not_authorized", false, false, ErrRuntimePinViolated)
	}
	if msg.req.RequireHealthyCapability && capView.HealthStatus != "healthy" {
		return nil, UnavailablePreflight(capView)
	}
	launchEnv := submitLaunchEnv(msg)
	if a.cfg.LaunchEnv != nil {
		a.cfg.LaunchEnv(msg.req, a.cfg.RuntimeID, launchEnv)
	}
	startReq := submitStartRequest(msg, capView, a.cfg.RuntimeID, launchEnv, a.cfg.Models, a.cfg.LaunchEnv != nil)
	spawn, err := adapter.BuildStart(startReq)
	if err != nil {
		return nil, submitBuildStartError(adapter, err)
	}
	driver, err := submitProtocolDriver(adapter, startReq)
	if err != nil {
		return nil, err
	}
	sess, err := a.startSubmitSession(msg, adapter, startReq, spawn, launchEnv, driver)
	if err != nil {
		return nil, err
	}
	handle := registerRunningSubmit(inFlight, msg.req.ID, string(msg.req.Provider), capView.CapabilityFingerprint, sess)
	watchSubmitCompletion(msg.req.ID, sess.Done(), a.stoppedCh, completeCh)
	return handle, nil
}
