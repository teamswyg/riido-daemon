package supervisor

import (
	"context"

	"github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) claimOne(ctx, claimCtx context.Context, rt *runtimeactor.Actor, status runtimeactor.Status, inFlight map[string]*runningTask) bool {
	req, err := a.cfg.Source.ClaimTask(claimCtx, status.RuntimeID)
	if err != nil || req == nil {
		return false
	}
	if req.ID == "" {
		return false
	}
	if _, dup := inFlight[req.ID]; dup {
		return false
	}
	report := reportContextFor(req)
	reportCtx := controlplane.ContextWithTaskReport(ctx, report)

	if err := a.cfg.Reporter.StartTask(reportCtx, req.ID); err != nil {
		return true
	}
	pin := runtimePinForClaim(status, req)
	taskCtx, cancel := context.WithCancel(ctx)
	inFlight[req.ID] = &runningTask{
		taskID:                req.ID,
		provider:              string(req.Provider),
		runtimeID:             status.RuntimeID,
		capabilityFingerprint: pin.capabilityFingerprint,
		nativeSessionID:       req.ResumeSessionID,
		ctx:                   taskCtx,
		report:                report,
		runtime:               rt,
		cancel:                cancel,
	}
	if nativeSessionInUse(inFlight, req.ID, string(req.Provider), req.ResumeSessionID) {
		a.finishTaskWithResult(ctx, inFlight, inFlight[req.ID], agentbridge.Result{Status: agentbridge.ResultBlocked, Error: "native session is already executing another turn", ExecutionContext: &assignment.ExecutionContext{DiagnosticCode: "session_scope_conflict", NextAction: "review_session", VerificationStatus: "not_configured"}})
		return true
	}
	eligibility := taskEligibility(status, req)
	if !eligibility.Eligible {
		a.finishTaskWithResult(ctx, inFlight, inFlight[req.ID], agentbridge.Result{
			Status: agentbridge.ResultBlocked,
			Error:  "supervisor: runtime ineligible: " + eligibility.Summary(),
		})
		return true
	}
	go a.forwardCancellation(taskCtx, req.ID, rt, cancel)
	go a.prepareAndSubmit(taskCtx, status, rt, req)
	return true
}

func nativeSessionInUse(inFlight map[string]*runningTask, taskID, provider, sessionID string) bool {
	if sessionID == "" {
		return false
	}
	for id, task := range inFlight {
		if id != taskID && task.provider == provider && task.nativeSessionID == sessionID {
			return true
		}
	}
	return false
}
