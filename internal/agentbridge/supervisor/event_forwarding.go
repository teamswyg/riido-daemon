package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) forwardSession(taskID string, events <-chan agentbridge.Event, results <-chan agentbridge.Result) {
	for ev := range events {
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

func (a *Actor) forwardCancellation(ctx context.Context, taskID string) {
	ch, err := a.cfg.Source.WatchCancellation(ctx, taskID)
	if err != nil {
		return
	}
	select {
	case cause, ok := <-ch:
		if !ok {
			return
		}
		select {
		case a.mailbox <- envelope{cancel: &cancelMsg{taskID: taskID, cause: cause}}:
		case <-a.stoppedCh:
		}
	case <-a.stoppedCh:
	case <-ctx.Done():
	}
}
