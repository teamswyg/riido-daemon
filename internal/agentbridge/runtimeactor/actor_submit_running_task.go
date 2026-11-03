package runtimeactor

import (
	"sync"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
)

func registerRunningSubmit(
	inFlight map[string]*runningTask,
	taskID string,
	provider string,
	capabilityFingerprint string,
	sess *session.Session,
) *SessionHandle {
	handle := &SessionHandle{TaskID: taskID, session: sess}
	inFlight[taskID] = &runningTask{
		taskID:                taskID,
		provider:              provider,
		capabilityFingerprint: capabilityFingerprint,
		handle:                handle,
	}
	return handle
}

func (a *Actor) contextCapabilityRelease(env map[string]string) func() {
	if a.cfg.LaunchEnv == nil || a.cfg.ReleaseContextCapability == nil {
		return func() {}
	}
	token := env["RIIDO_RUN_TOKEN"]
	return sync.OnceFunc(func() { a.cfg.ReleaseContextCapability(token) })
}

func watchSubmitCompletion(taskID string, doneCh, stopped <-chan struct{}, completeCh chan<- string, release func()) {
	go func() {
		defer release()
		select {
		case <-doneCh:
		case <-stopped:
			return
		}
		release()
		select {
		case completeCh <- taskID:
		case <-stopped:
		}
	}()
}
