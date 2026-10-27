package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

type claimReadyBlockReporter struct {
	*reporterProbe
	entered chan struct{}
}

func (r *claimReadyBlockReporter) StartTask(ctx context.Context, _ string) error {
	close(r.entered)
	<-ctx.Done()
	return ctx.Err()
}

func TestClaimProofPreparationReservedBeforeGoroutine(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	source := &heartbeatDuringPrepareSource{req: bridge.TaskRequest{ID: "claim", Provider: "fake", LocalExecutionScope: scope}}
	reporter := &claimReadyBlockReporter{reporterProbe: newReporterProbe(), entered: make(chan struct{})}
	actor := &Actor{cfg: Config{Source: source, Reporter: reporter}, mailbox: make(chan envelope, 1), stoppedCh: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inFlight := map[string]*runningTask{}
	if !actor.claimOne(ctx, ctx, nil, runtimeactor.Status{RuntimeID: "runtime"}, inFlight) {
		t.Fatal("fixture claim failed")
	}
	select {
	case <-reporter.entered:
	case <-time.After(time.Second):
		t.Fatal("Ready worker did not enter")
	}
	if _, ok := scope.Tracker().Seal(); ok {
		t.Fatal("preparation worker lacked pre-launch reservation")
	}
	cancel()
	select {
	case <-actor.mailbox:
	case <-time.After(time.Second):
		t.Fatal("preparation worker did not exit")
	}
	if _, ok := scope.Tracker().Seal(); ok {
		t.Fatal("late preparation completion upgraded unknown")
	}
}
