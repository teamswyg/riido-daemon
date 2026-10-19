package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type cancellationFlowSource struct {
	*controlplane.MemorySource
	watching chan string
}

func (s *cancellationFlowSource) WatchCancellation(ctx context.Context, id string) (<-chan error, error) {
	watch, err := s.MemorySource.WatchCancellation(ctx, id)
	if err == nil {
		s.watching <- id
	}
	return watch, err
}

type cancellationFlowReporter struct {
	failed  chan struct{}
	results chan agentbridge.Result
}

func (*cancellationFlowReporter) StartTask(context.Context, string) error { return nil }
func (r *cancellationFlowReporter) ReportEvent(_ context.Context, id string, ev agentbridge.Event) error {
	if id == "first" && ev.Kind == agentbridge.EventLifecycle && ev.Phase == agentbridge.StateRunning {
		select {
		case r.failed <- struct{}{}:
		default:
		}
		return errors.New("Running delivery unavailable before commit")
	}
	return nil
}

func (r *cancellationFlowReporter) CompleteTask(_ context.Context, _ string, res agentbridge.Result) error {
	r.results <- res
	return nil
}

func TestCancellationSettlesOwnershipAndAdmitsNextTask(t *testing.T) {
	p := &retryStartProcess{started: make(chan *process.FakeRunning, 2)}
	runtime, _ := newRecoveryRuntime(t, "fake", "", p)
	source := &cancellationFlowSource{MemorySource: controlplane.NewMemorySource(), watching: make(chan string, 2)}
	reporter := &cancellationFlowReporter{failed: make(chan struct{}, 1), results: make(chan agentbridge.Result, 2)}
	first := recoveryRequest()
	first.ID = "first"
	source.Enqueue(first)
	actor := startRoutingSupervisor(t, Config{DaemonID: "test", Runtime: runtime, Source: source, Reporter: reporter, PollEvery: 5 * time.Millisecond})
	awaitRecoverySpawn(t, p)
	select {
	case <-source.watching:
	case <-time.After(time.Second):
		t.Fatal("cancellation watch missing")
	}
	select {
	case <-reporter.failed:
	case <-time.After(time.Second):
		t.Fatal("Running delivery did not fail")
	}
	source.Cancel(first.ID, context.Canceled)
	select {
	case result := <-reporter.results:
		if result.Status != agentbridge.ResultCancelled {
			t.Fatal(result.Status)
		}
	case <-time.After(time.Second):
		t.Fatal("pending Running blocked cancellation")
	}
	second := recoveryRequest()
	second.ID = "second"
	source.Enqueue(second)
	run := awaitRecoverySpawn(t, p)
	for _, ids := range actor.OwnedExecutions() {
		for _, id := range ids {
			if id == first.ID {
				t.Fatal("cancelled execution retained heartbeat ownership")
			}
		}
	}
	completeTaskResultProcess(run)
	select {
	case result := <-reporter.results:
		if result.Status != agentbridge.ResultCompleted {
			t.Fatal(result.Status)
		}
	case <-time.After(time.Second):
		t.Fatal("next task completion blocked")
	}
	if p.count() != 2 {
		t.Fatalf("dispatches=%d; want one per task", p.count())
	}
}
