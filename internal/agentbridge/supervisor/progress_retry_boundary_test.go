package supervisor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type criticalBoundaryReporter struct {
	*reporterProbe
	progressExited <-chan struct{}
}

func (r *criticalBoundaryReporter) ReportEvent(ctx context.Context, id string, ev agentbridge.Event) error {
	select {
	case <-r.progressExited:
		return r.reporterProbe.ReportEvent(ctx, id, ev)
	default:
		return errors.New("critical retry overlapped progress")
	}
}

func TestCriticalRetryPreemptsProgressDelivery(t *testing.T) {
	entered, exited := make(chan struct{}), make(chan struct{})
	delivery := newProgressDelivery(t.Context(), func(ctx context.Context, _ agentbridge.Event) error {
		close(entered)
		<-ctx.Done()
		close(exited)
		return ctx.Err()
	})
	defer delivery.stop()
	delivery.enqueue(agentbridge.Event{Kind: agentbridge.EventProgress, Text: "still working"})
	<-entered
	reporter := &criticalBoundaryReporter{reporterProbe: newReporterProbe(), progressExited: exited}
	a := &Actor{cfg: Config{Reporter: reporter}}
	task := &runningTask{taskID: "claim", delivery: delivery, pendingEvents: []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}}}
	if !a.flushNextEventReport(t.Context(), task) || len(task.pendingEvents) != 0 {
		t.Fatal("critical retry must cancel progress before reporting")
	}
}
