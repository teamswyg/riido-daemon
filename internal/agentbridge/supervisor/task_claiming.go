package supervisor

import (
	"context"

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

	_ = a.cfg.Reporter.StartTask(reportCtx, req.ID)
	eligibility := taskEligibility(status, req)
	if !eligibility.Eligible {
		_ = a.cfg.Reporter.CompleteTask(reportCtx, req.ID, agentbridge.Result{
			Status: agentbridge.ResultBlocked,
			Error:  "supervisor: runtime ineligible: " + eligibility.Summary(),
		})
		return true
	}
	taskCtx, cancel := context.WithCancel(ctx)
	inFlight[req.ID] = &runningTask{taskID: req.ID, ctx: taskCtx, report: report, runtime: rt, cancel: cancel}
	go a.forwardCancellation(taskCtx, req.ID)
	go a.prepareAndSubmit(taskCtx, status, rt, req)
	return true
}
