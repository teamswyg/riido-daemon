package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (a *Actor) reportTaskEvent(ctx context.Context, task *runningTask, ev agentbridge.Event) {
	if task == nil {
		return
	}
	if claimDeliveryCancelled(task) && obsoleteAfterCancellation(ev) {
		a.recordSupersededNotification(ctx, task, ev)
		return
	}
	reportCtx := controlplane.ContextWithTaskReport(ctx, task.report)
	err := task.timing.report(func() error { return a.cfg.Reporter.ReportEvent(reportCtx, task.taskID, ev) })
	if err != nil {
		retainReportEvent(task, ev)
	}
}

func retainReportEvent(task *runningTask, ev agentbridge.Event) {
	if task == nil || !shouldRetainReportEvent(ev) || len(task.pendingEvents) >= maxPendingReportEvents {
		return
	}
	task.pendingEvents = append(task.pendingEvents, ev)
}

func (a *Actor) retryEventReports(ctx context.Context, inFlight map[string]*runningTask) bool {
	reported := false
	for _, task := range inFlight {
		a.supersedeCancelledNotifications(ctx, task)
		if len(task.pendingEvents) == 0 {
			continue
		}
		if a.flushNextEventReport(ctx, task) {
			reported = true
		}
	}
	return reported
}

func (a *Actor) flushNextEventReport(ctx context.Context, task *runningTask) bool {
	reportCtx := controlplane.ContextWithTaskReport(ctx, task.report)
	if err := task.timing.report(func() error { return a.cfg.Reporter.ReportEvent(reportCtx, task.taskID, task.pendingEvents[0]) }); err != nil {
		return false
	}
	task.pendingEvents = task.pendingEvents[1:]
	return true
}
