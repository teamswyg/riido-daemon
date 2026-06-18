package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type reporterProbe struct {
	started chan string
	events  chan agentbridge.Event
	results chan agentbridge.Result
}

func newReporterProbe() *reporterProbe {
	return &reporterProbe{
		started: make(chan string, 4),
		events:  make(chan agentbridge.Event, 8),
		results: make(chan agentbridge.Result, 4),
	}
}

func (r *reporterProbe) StartTask(_ context.Context, taskID string) error {
	r.started <- taskID
	return nil
}

func (r *reporterProbe) ReportEvent(_ context.Context, _ string, ev agentbridge.Event) error {
	r.events <- ev
	return nil
}

func (r *reporterProbe) CompleteTask(_ context.Context, _ string, res agentbridge.Result) error {
	r.results <- res
	return nil
}
