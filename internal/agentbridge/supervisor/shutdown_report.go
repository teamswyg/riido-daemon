package supervisor

import (
	"context"
	"time"
)

const shutdownRetryEvery = 50 * time.Millisecond

func (a *Actor) drainShutdownReports(
	ctx context.Context,
	inFlight map[string]*runningTask,
	detachedReports *[]detachedReport,
) {
	for hasPendingShutdownReport(inFlight, detachedReports) {
		a.retryEventReports(ctx, inFlight)
		a.retryTerminalReports(ctx, inFlight)
		a.retryDetachedReports(ctx, detachedReports)
		if !hasPendingShutdownReport(inFlight, detachedReports) || waitShutdownRetry(ctx) {
			return
		}
	}
}

func hasPendingShutdownReport(
	inFlight map[string]*runningTask,
	detachedReports *[]detachedReport,
) bool {
	return len(inFlight) > 0 || detachedReports != nil && len(*detachedReports) > 0
}

func waitShutdownRetry(ctx context.Context) bool {
	timer := time.NewTimer(shutdownRetryEvery)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return true
	case <-timer.C:
		return false
	}
}
