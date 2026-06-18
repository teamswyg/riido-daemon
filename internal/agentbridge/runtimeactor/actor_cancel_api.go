package runtimeactor

import "context"

// Cancel asks the actor to cancel an in-flight task.
func (a *Actor) Cancel(ctx context.Context, taskID, reason string) error {
	reply := make(chan error, 1)
	select {
	case a.mailbox <- envelope{cancel: &cancelMsg{ctx: ctx, taskID: taskID, reason: reason, reply: reply}}:
	case <-a.stoppedCh:
		return ErrActorStopped
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-a.stoppedCh:
		return ErrActorStopped
	case <-ctx.Done():
		return ctx.Err()
	}
}
