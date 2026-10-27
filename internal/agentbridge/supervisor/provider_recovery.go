package supervisor

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/providerrecovery"
)

type recoveryFailure struct {
	cause   error
	blocked bool
}

func (e *recoveryFailure) Error() string {
	if e.cause != nil {
		return e.cause.Error()
	}
	return "provider recovery stopped"
}
func (e *recoveryFailure) Unwrap() error { return e.cause }

func (a *Actor) recoverAndSubmit(ctx context.Context, status runtimeactor.Status, rt *runtimeactor.Actor, req *bridge.TaskRequest, work *executionproof.HostWork) {
	policy, err := req.RecoveryPolicy.Normalized()
	circuits := a.recovery
	if circuits == nil {
		circuits = providerrecovery.NewCircuits()
	}
	observation := assignment.RecoveryContext{State: "checking", Reason: "none", Provider: string(req.Provider), Circuit: "closed", OutcomeCertainty: "not_started"}
	var prepared *preparedWorkspace
	workspacePrepared := false
	scope := ""
	finish := func(cause error, blocked bool) {
		if observation.Circuit == "half_open" {
			if ctx.Err() != nil {
				circuits.Release(scope, req.ID)
			} else {
				circuits.Complete(scope, req.ID, false, time.Now())
			}
			observation.Circuit = "open"
		}
		if ctx.Err() != nil {
			observation.State = "cancelled"
			observation.Reason = "cancelled"
			cause = ctx.Err()
		}
		copyObservation := observation.Sanitized()
		a.forwardActivation(taskActivationMsg{hostWork: work, claimContext: ctx, taskID: req.ID, prepared: prepared, runtime: rt, status: status, recovery: &copyObservation, recoveryScope: scope, err: &recoveryFailure{cause: cause, blocked: blocked}})
	}
	if err != nil {
		observation.State = "user_action_required"
		observation.Reason = "not_authorized"
		finish(err, true)
		return
	}
	budgetCtx, cancel := context.WithTimeout(ctx, time.Duration(policy.MaxElapsedSeconds)*time.Second)
	defer cancel()

	for observation.Attempt < policy.MaxAttempts {
		if budgetCtx.Err() != nil {
			observation.State = "unavailable"
			observation.Reason = "deadline_exceeded"
			finish(budgetCtx.Err(), false)
			return
		}
		observation.Attempt++
		status, err = rt.Status(budgetCtx)
		if err != nil {
			observation.State = "user_action_required"
			observation.Reason = "unknown_outcome"
			finish(err, true)
			return
		}
		capability, found := findCapability(status.Capabilities, string(req.Provider))
		scope = string(req.Provider) + ":" + capability.CapabilityFingerprint
		circuit, wait := circuits.Acquire(scope, req.ID, time.Now())
		observation.Circuit = circuit
		if wait > 0 {
			observation.State = "backoff"
			observation.Reason = "transient_preflight"
			if err = a.reportRecovery(budgetCtx, req, observation); err != nil {
				observation.State = "user_action_required"
				observation.Reason = "not_authorized"
				finish(err, true)
				return
			}
			if !recoveryWait(budgetCtx, wait) {
				observation.State = "unavailable"
				observation.Reason = "deadline_exceeded"
				finish(budgetCtx.Err(), false)
				return
			}
		}
		if wait > 0 {
			observation.Circuit, wait = circuits.Acquire(scope, req.ID, time.Now())
			if wait > 0 {
				continue
			}
		}
		observation.State = "checking"
		observation.Provider = string(req.Provider)
		if err = a.reportRecovery(budgetCtx, req, observation); err != nil {
			observation.State = "user_action_required"
			observation.Reason = "not_authorized"
			finish(err, true)
			return
		}
		if budgetCtx.Err() != nil {
			finish(budgetCtx.Err(), false)
			return
		}
		eligibility := taskEligibility(status, req)
		if !eligibility.Eligible {
			if found && !capability.Available {
				err = runtimeactor.UnavailablePreflight(capability)
			} else {
				observation.State = "user_action_required"
				observation.Reason = "not_authorized"
				finish(fmt.Errorf("supervisor: runtime ineligible: %s", eligibility.Summary()), true)
				return
			}
		} else {
			if !workspacePrepared {
				req.ExpectedCapabilityFingerprint = capability.CapabilityFingerprint
				// Repository materialization is ordinary task preparation, not a
				// provider recovery attempt. Pause only elapsed recovery accounting:
				// keep cancellation and resume with exactly the remaining budget.
				deadline, _ := budgetCtx.Deadline()
				remaining := time.Until(deadline)
				if remaining <= 0 {
					observation.State = "unavailable"
					observation.Reason = "deadline_exceeded"
					finish(context.DeadlineExceeded, false)
					return
				}
				cancel()
				prepared, err = a.prepareWorkspace(ctx, status, req)
				nextBudgetCtx, releaseBudget := context.WithTimeout(ctx, remaining)
				defer releaseBudget()
				budgetCtx = nextBudgetCtx
				if err != nil {
					finish(fmt.Errorf("%w: %w", errWorkspacePreparation, err), false)
					return
				}
				workspacePrepared = true
			}
			if prepared != nil {
				snapshot := *req
				prepared.request = &snapshot
			}
			var handle *runtimeactor.SessionHandle
			handle, err = rt.SubmitPreflight(budgetCtx, ctx, *req)
			if err == nil && handle != nil {
				observation.State = "dispatched"
				observation.Reason = "none"
				observation.OutcomeCertainty = "started"
				copyObservation := observation.Sanitized()
				a.forwardActivation(taskActivationMsg{hostWork: work, claimContext: ctx, taskID: req.ID, prepared: prepared, handle: handle, runtime: rt, status: status, recovery: &copyObservation, recoveryScope: scope})
				return
			}
		}
		var preflight *runtimeactor.PreflightError
		if !errors.As(err, &preflight) {
			observation.State = "user_action_required"
			observation.Reason = "unknown_outcome"
			observation.OutcomeCertainty = "unknown"
			finish(err, true)
			return
		}
		observation.Reason = preflight.Reason
		circuits.Failure(scope, req.ID, time.Now())
		if observation.Attempt >= policy.MaxAttempts {
			observation.State = "unavailable"
			observation.Reason = "budget_exhausted"
			finish(err, false)
			return
		}
		if preflight.Switchable && observation.SwitchCount < policy.MaxSwitches {
			nextRT, nextStatus, nextReq, switchErr := a.recoveryFallback(budgetCtx, status, req, policy.Fallbacks)
			if switchErr == nil {
				observation.State = "switching"
				observation.Provider = string(nextReq.Provider)
				observation.SwitchCount++
				if err = a.reportRecovery(budgetCtx, req, observation); err != nil {
					observation.State = "user_action_required"
					observation.Reason = "not_authorized"
					finish(err, true)
					return
				}
				if err = a.reconfigureRecoveryWorkspace(nextStatus, nextReq, prepared); err != nil {
					observation.State = "user_action_required"
					observation.Reason = "not_authorized"
					finish(err, true)
					return
				}
				req = nextReq
				rt = nextRT
				status = nextStatus
				continue
			}
		}
		if !preflight.Retryable {
			observation.State = "user_action_required"
			finish(err, true)
			return
		}
		observation.State = "backoff"
		if err = a.reportRecovery(budgetCtx, req, observation); err != nil {
			observation.State = "user_action_required"
			observation.Reason = "not_authorized"
			finish(err, true)
			return
		}
		if !recoveryWait(budgetCtx, 200*time.Millisecond+time.Duration(rand.IntN(100))*time.Millisecond) {
			observation.State = "unavailable"
			observation.Reason = "deadline_exceeded"
			finish(budgetCtx.Err(), false)
			return
		}
		if err = rt.RefreshCapability(budgetCtx, string(req.Provider)); err != nil {
			observation.State = "unavailable"
			finish(err, false)
			return
		}
	}
	observation.State = "unavailable"
	observation.Reason = "budget_exhausted"
	finish(errors.New("provider recovery budget exhausted"), false)
}

func recoveryWait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return ctx.Err() == nil
	case <-ctx.Done():
		return false
	}
}

func (a *Actor) reportRecovery(ctx context.Context, req *bridge.TaskRequest, c assignment.RecoveryContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c = c.Sanitized()
	return a.cfg.Reporter.ReportEvent(controlplane.ContextWithTaskReport(ctx, reportContextFor(req)), req.ID, agentbridge.Event{Kind: agentbridge.EventProgress, Text: "Provider recovery: " + c.State, RecoveryContext: &c})
}
