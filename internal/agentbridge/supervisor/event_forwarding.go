package supervisor

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) forwardSession(taskID string, events <-chan agentbridge.Event, results <-chan agentbridge.Result) {
	for ev := range events {
		a.cancelCurrentClaim()
		select {
		case a.mailbox <- envelope{taskEvent: &taskEventMsg{taskID: taskID, event: ev}}:
		case <-a.stoppedCh:
			return
		}
	}
	res, ok := <-results
	if !ok {
		return
	}
	select {
	case a.mailbox <- envelope{taskResult: &taskResultMsg{taskID: taskID, result: res}}:
	case <-a.stoppedCh:
	}
}

func (a *Actor) forwardCancellation(ctx context.Context, taskID string, rt *runtimeactor.Actor, cancel context.CancelFunc) {
	ch, err := a.cfg.Source.WatchCancellation(ctx, taskID)
	if err != nil {
		return
	}
	select {
	case cause, ok := <-ch:
		if !ok {
			return
		}
		cancel()
		a.cancelCurrentClaim()
		requestCtx, release := context.WithTimeout(context.Background(), 5*time.Second)
		reason := "cancelled"
		if cause != nil {
			reason = cause.Error()
		}
		_ = rt.Cancel(requestCtx, taskID, reason)
		release()
		select {
		case a.mailbox <- envelope{cancel: &cancelMsg{taskID: taskID, cause: cause}}:
		case <-a.stoppedCh:
		}
	case <-a.stoppedCh:
	case <-ctx.Done():
	}
}
