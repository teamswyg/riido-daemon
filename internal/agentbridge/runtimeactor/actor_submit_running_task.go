package runtimeactor

import "github.com/teamswyg/riido-daemon/internal/agentbridge/session"

func registerRunningSubmit(
	inFlight map[string]*runningTask,
	taskID string,
	provider string,
	sess *session.Session,
) *SessionHandle {
	handle := &SessionHandle{TaskID: taskID, session: sess}
	inFlight[taskID] = &runningTask{
		taskID:   taskID,
		provider: provider,
		handle:   handle,
	}
	return handle
}

func watchSubmitCompletion(taskID string, doneCh, stopped <-chan struct{}, completeCh chan<- string) {
	go func() {
		select {
		case <-doneCh:
		case <-stopped:
			return
		}
		select {
		case completeCh <- taskID:
		case <-stopped:
		}
	}()
}
