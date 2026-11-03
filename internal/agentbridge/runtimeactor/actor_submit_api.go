package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// Submit posts a TaskRequest to the actor. The stoppedCh wait guard prevents
// callers from blocking after a buffered mailbox send races with actor stop.
func (a *Actor) Submit(ctx context.Context, req bridge.TaskRequest) (*SessionHandle, error) {
	return a.SubmitPreflight(ctx, ctx, req)
}

// SubmitPreflight bounds admission separately from the authorized run lifetime.
func (a *Actor) SubmitPreflight(ctx, runCtx context.Context, req bridge.TaskRequest) (*SessionHandle, error) {
	return a.SubmitWithNativeResume(ctx, runCtx, req, nil)
}

func (a *Actor) SubmitWithNativeResume(ctx, runCtx context.Context, req bridge.TaskRequest, selected *NativeResumeRuntime) (*SessionHandle, error) {
	reply := make(chan submitReply, 1)
	select {
	case a.mailbox <- envelope{submit: &submitMsg{ctx: ctx, runCtx: runCtx, req: req, reply: reply, nativeResume: selected}}:
	case <-a.stoppedCh:
		return nil, ErrActorStopped
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.handle, res.err
	case <-a.stoppedCh:
		return nil, ErrActorStopped
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// SubmitContinuation preserves the executable selected for a completed run.
func (a *Actor) SubmitContinuation(ctx context.Context, req bridge.TaskRequest, previous *SessionHandle) (*SessionHandle, error) {
	var selected *NativeResumeRuntime
	if previous != nil && previous.TaskID == req.ID {
		selected = previous.nativeResume
	}
	return a.SubmitWithNativeResume(ctx, ctx, req, selected)
}
