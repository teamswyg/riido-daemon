package supervisor

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type stateEventRetryReporter struct {
	*reporterProbe
	failures int
	attempts chan int
	count    int
}

func newStateEventRetryReporter(failures int) *stateEventRetryReporter {
	return &stateEventRetryReporter{
		reporterProbe: newReporterProbe(),
		failures:      failures,
		attempts:      make(chan int, failures+2),
	}
}

func (r *stateEventRetryReporter) ReportEvent(
	ctx context.Context,
	taskID string,
	ev agentbridge.Event,
) error {
	if !shouldRetainReportEvent(ev) {
		return r.reporterProbe.ReportEvent(ctx, taskID, ev)
	}
	r.count++
	r.attempts <- r.count
	if r.count <= r.failures {
		return errors.New("state event report rejected")
	}
	return r.reporterProbe.ReportEvent(ctx, taskID, ev)
}
