package runtimeactor

import (
	"errors"
	"fmt"
)

func (a *Actor) handleCancel(inFlight map[string]*runningTask, msg *cancelMsg) error {
	task, ok := inFlight[msg.taskID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownTask, msg.taskID)
	}
	cause := errors.New(msg.reason)
	if msg.reason == "" {
		cause = errors.New("cancelled")
	}
	task.handle.session.CancelWithContext(msg.ctx, cause)
	return nil
}
