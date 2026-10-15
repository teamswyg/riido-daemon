package runtimeactor

import "context"

// RefreshCapability schedules a single background probe. A subsequent Submit
// waits for that probe outside the actor's lifecycle mailbox.
func (a *Actor) RefreshCapability(ctx context.Context, provider string) error {
	reply := make(chan error, 1)
	select {
	case a.mailbox <- envelope{refresh: &refreshMsg{provider: provider, reply: reply}}:
	case <-ctx.Done():
		return ctx.Err()
	case <-a.stoppedCh:
		return ErrActorStopped
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-a.stoppedCh:
		return ErrActorStopped
	}
}
