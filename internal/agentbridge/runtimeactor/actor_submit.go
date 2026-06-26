package runtimeactor

import (
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
	if err := a.validateSubmit(inFlight, msg); err != nil {
		return nil, err
	}
	adapter, capView, err := a.submitCapability(msg, adapters, caps, detectedAt)
	if err != nil {
		return nil, err
	}
	launchEnv := submitLaunchEnv(msg)
	startReq := submitStartRequest(msg, capView, a.cfg.RuntimeID, launchEnv)
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
