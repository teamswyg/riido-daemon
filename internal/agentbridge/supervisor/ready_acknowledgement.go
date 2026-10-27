package supervisor

import (
	"context"
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

const (
	readyMaxAttempts = 3
	readyMaxElapsed  = 30 * time.Second
)

type readyAcknowledgementFailure struct {
	cause   error
	blocked bool
}

func (e *readyAcknowledgementFailure) Error() string {
	return "assignment Ready acknowledgement was not confirmed; execution did not start"
}
func (e *readyAcknowledgementFailure) Unwrap() error { return e.cause }

func (a *Actor) prepareClaimedTask(ctx context.Context, status runtimeactor.Status, rt *runtimeactor.Actor, req *bridge.TaskRequest, report controlplane.TaskReportContext, work *executionproof.HostWork) {
	reportCtx := controlplane.ContextWithTaskReport(ctx, report)
	if err := a.acknowledgeReady(reportCtx, req.ID); err != nil {
		a.forwardActivation(taskActivationMsg{hostWork: work, claimContext: ctx, taskID: req.ID, err: err})
		return
	}
	if err := ctx.Err(); err != nil {
		a.forwardActivation(taskActivationMsg{hostWork: work, claimContext: ctx, taskID: req.ID, err: err})
		return
	}
	a.prepareAndSubmit(ctx, status, rt, req, work)
}

func (a *Actor) acknowledgeReady(ctx context.Context, taskID string) error {
	budgetCtx, cancel := context.WithTimeout(ctx, readyMaxElapsed)
	defer cancel()
	for attempt := 1; attempt <= readyMaxAttempts; attempt++ {
		if err := budgetCtx.Err(); err != nil {
			return &readyAcknowledgementFailure{cause: err}
		}
		err := a.cfg.Reporter.StartTask(budgetCtx, taskID)
		if budgetCtx.Err() != nil {
			return &readyAcknowledgementFailure{cause: budgetCtx.Err()}
		}
		if err == nil {
			return nil
		}
		var retryable interface{ RetryableReady() bool }
		retry := errors.As(err, &retryable) && retryable.RetryableReady()
		if !retry || attempt == readyMaxAttempts {
			return &readyAcknowledgementFailure{cause: err, blocked: !retry}
		}
		if !recoveryWait(budgetCtx, time.Duration(attempt)*200*time.Millisecond) {
			return &readyAcknowledgementFailure{cause: budgetCtx.Err()}
		}
	}
	return &readyAcknowledgementFailure{cause: context.DeadlineExceeded}
}
