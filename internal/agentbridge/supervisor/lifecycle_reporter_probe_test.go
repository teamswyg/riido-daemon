package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

type lifecycleReporterProbe struct {
	*reporterProbe
	completeLevels chan lifecycle.ShutdownLevel
}

func newLifecycleReporterProbe() *lifecycleReporterProbe {
	return &lifecycleReporterProbe{
		reporterProbe:  newReporterProbe(),
		completeLevels: make(chan lifecycle.ShutdownLevel, 4),
	}
}

func (r *lifecycleReporterProbe) CompleteTask(ctx context.Context, taskID string, res agentbridge.Result) error {
	r.completeLevels <- lifecycle.FromContext(ctx).ShutdownLevel()
	return r.reporterProbe.CompleteTask(ctx, taskID, res)
}
