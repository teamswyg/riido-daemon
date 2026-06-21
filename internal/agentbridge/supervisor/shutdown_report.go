package supervisor

import (
	"context"
	"time"
)

const shutdownReportRetryEvery = 50 * time.Millisecond

func (a *Actor) drainShutdownReports(ctx context.Context, inFlight map[string]*runningTask) {
	for len(inFlight) > 0 {
		a.retryEventReports(ctx, inFlight)
		a.retryTerminalReports(ctx, inFlight)
		if len(inFlight) == 0 || waitShutdownReportRetry(ctx) {
			return
		}
	}
}

func waitShutdownReportRetry(ctx context.Context) bool {
	timer := time.NewTimer(shutdownReportRetryEvery)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return true
	case <-timer.C:
		return false
	}
}
