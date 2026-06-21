package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const maxDetachedReports = 16

type detachedReport struct {
	taskID string
	event  *agentbridge.Event
	result *agentbridge.Result
}

func detachedEvent(taskID string, ev agentbridge.Event) detachedReport {
	return detachedReport{taskID: taskID, event: &ev}
}

func detachedResult(taskID string, res agentbridge.Result) detachedReport {
	return detachedReport{taskID: taskID, result: &res}
}

func (a *Actor) reportOrRetainDetached(
	ctx context.Context,
	pending *[]detachedReport,
	report detachedReport,
) {
	if a.reportDetached(ctx, report) {
		return
	}
	retainDetachedReport(pending, report)
}

func retainDetachedReport(pending *[]detachedReport, report detachedReport) {
	if pending == nil || !shouldRetainDetachedReport(report) || len(*pending) >= maxDetachedReports {
		return
	}
	*pending = append(*pending, report)
}

func shouldRetainDetachedReport(report detachedReport) bool {
	if report.event != nil {
		return shouldRetainReportEvent(*report.event)
	}
	return report.result != nil
}

func (a *Actor) retryDetachedReports(ctx context.Context, pending *[]detachedReport) bool {
	if pending == nil || len(*pending) == 0 {
		return false
	}
	if !a.reportDetached(ctx, (*pending)[0]) {
		return false
	}
	*pending = (*pending)[1:]
	return true
}

func (a *Actor) reportDetached(ctx context.Context, report detachedReport) bool {
	switch {
	case report.event != nil:
		return a.cfg.Reporter.ReportEvent(ctx, report.taskID, *report.event) == nil
	case report.result != nil:
		return a.cfg.Reporter.CompleteTask(ctx, report.taskID, *report.result) == nil
	default:
		return true
	}
}
