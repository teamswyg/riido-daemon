package supervisor

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type terminalRetryReporter struct {
	*reporterProbe
	failures int
	attempts chan int
	count    int
}

func newTerminalRetryReporter(failures int) *terminalRetryReporter {
	return &terminalRetryReporter{
		reporterProbe: newReporterProbe(),
		failures:      failures,
		attempts:      make(chan int, failures+2),
	}
}

func (r *terminalRetryReporter) CompleteTask(
	ctx context.Context,
	taskID string,
	res agentbridge.Result,
) error {
	r.count++
	r.attempts <- r.count
	if r.count <= r.failures {
		return errors.New("terminal report rejected")
	}
	return r.reporterProbe.CompleteTask(ctx, taskID, res)
}
