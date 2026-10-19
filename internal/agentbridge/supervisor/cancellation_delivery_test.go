package supervisor

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/execution"
)

type cancellationDeliveryReporter struct {
	failEvent    bool
	failTerminal bool
	cancelling   bool
	order        []string
	results      []agentbridge.Result
	accepted     []agentbridge.Result
}

func (*cancellationDeliveryReporter) StartTask(context.Context, string) error { return nil }
func (r *cancellationDeliveryReporter) ReportEvent(_ context.Context, _ string, ev agentbridge.Event) error {
	if r.failEvent {
		r.failEvent = false
		r.order = append(r.order, "event:failed")
		return errors.New("transport retries exhausted before commit")
	}
	if r.cancelling && obsoleteAfterCancellation(ev) {
		r.order = append(r.order, "event:rejected")
		return errors.New("invalid assignment transition from cancelling")
	}
	r.order = append(r.order, "event:accepted")
	return nil
}

func (r *cancellationDeliveryReporter) CompleteTask(_ context.Context, _ string, res agentbridge.Result) error {
	r.results = append(r.results, res)
	if r.failTerminal {
		r.failTerminal = false
		r.order = append(r.order, "terminal:unknown")
		return errors.New("terminal acknowledgement was lost")
	}
	if r.cancelling && res.Status == agentbridge.ResultCompleted {
		return errors.New("invalid assignment transition cancelling -> completed")
	}
	r.order = append(r.order, "terminal:accepted")
	r.accepted = append(r.accepted, res)
	return nil
}

func cancellationDeliveryFixture(reporter *cancellationDeliveryReporter) (*Actor, *runningTask, map[string]*runningTask) {
	ctx, cancel := context.WithCancel(context.Background())
	task := &runningTask{taskID: "execution", runtimeID: "rt:fake", ctx: ctx, cancel: cancel}
	return &Actor{cfg: Config{Reporter: reporter}}, task, map[string]*runningTask{task.taskID: task}
}

func cancelDeliveryClaim(a *Actor, task *runningTask, tasks map[string]*runningTask) {
	a.handleTaskCancel(context.Background(), &cancelMsg{taskID: task.taskID, claimContext: task.ctx, cause: context.Canceled}, tasks)
}

func deliverCancellationTestResult(a *Actor, task *runningTask, tasks map[string]*runningTask, res agentbridge.Result) {
	detached := []detachedReport{}
	a.handleTaskResult(context.Background(), &taskResultMsg{taskID: task.taskID, result: res}, tasks, &detached)
}

func TestCancelledClaimSupersedesPendingLifecycleBeforeTerminal(t *testing.T) {
	events := []agentbridge.Event{
		{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning},
		{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateToolRunning},
		{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateWaitingUserInput},
		{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateWaitingToolApproval},
		{Kind: agentbridge.EventToolApprovalNeeded, Tool: agentbridge.ToolRef{Name: "mock-tool", ID: "mock-approval"}},
	}
	for _, ev := range events {
		t.Run(string(ev.Kind)+":"+string(ev.Phase), func(t *testing.T) {
			reporter := &cancellationDeliveryReporter{failEvent: true}
			a, task, tasks := cancellationDeliveryFixture(reporter)
			a.reportTaskEvent(context.Background(), task, ev)
			if len(task.pendingEvents) != 1 {
				t.Fatal("failed event not retained")
			}
			reporter.cancelling = true
			cancelDeliveryClaim(a, task, tasks)
			deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCancelled})
			for range 3 {
				a.retryEventReports(context.Background(), tasks)
				a.retryTerminalReports(context.Background(), tasks)
			}
			a.publishHeartbeatExecutions(tasks)
			if len(reporter.accepted) != 1 || reporter.accepted[0].Status != agentbridge.ResultCancelled || len(tasks) != 0 || len(a.OwnedExecutions()) != 0 {
				t.Fatalf("cancelled terminal blocked: events=%d tasks=%d results=%+v order=%v", len(task.pendingEvents), len(tasks), reporter.results, reporter.order)
			}
			if len(task.supersededEvents) != 1 || !reflect.DeepEqual(task.supersededEvents[0], ev) {
				t.Fatal("superseded event evidence missing")
			}
		})
	}
}

func TestCancellationBeforeQueuedCompletionWins(t *testing.T) {
	for _, hostChecks := range []bool{false, true} {
		t.Run(map[bool]string{false: "without-host-checks", true: "with-host-checks"}[hostChecks], func(t *testing.T) {
			reporter := &cancellationDeliveryReporter{cancelling: true}
			a, task, tasks := cancellationDeliveryFixture(reporter)
			if hostChecks {
				task.execution = &execution.Prepared{Binding: execution.Binding{ChangedFiles: []string{"not-created"}, Checks: [][]string{{"must-not-execute-review-check"}}}}
			}
			cancelDeliveryClaim(a, task, tasks)
			deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "provider answer finished before delivery", SessionID: "native", Usage: agentbridge.Usage{PromptTokens: 7}})
			if len(reporter.accepted) != 1 || reporter.accepted[0].Status != agentbridge.ResultCancelled || len(tasks) != 0 || task.verificationPending {
				t.Fatalf("prior cancel lost: results=%+v remaining=%d verification=%v", reporter.results, len(tasks), task.verificationPending)
			}
			if reporter.accepted[0].Usage.PromptTokens != 7 || reporter.accepted[0].SessionID != "native" {
				t.Fatal("cancellation discarded provenance/usage")
			}
		})
	}
}

func TestInternalFinishCancellationDoesNotRewriteCompletedRetry(t *testing.T) {
	reporter := &cancellationDeliveryReporter{failTerminal: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "answer"})
	if task.ctx.Err() == nil || task.cancelCause != nil {
		t.Fatal("expected only internal finish cancellation")
	}
	a.retryTerminalReports(context.Background(), tasks)
	if len(reporter.results) != 2 || !reflect.DeepEqual(reporter.results[0], reporter.results[1]) || len(tasks) != 0 || reporter.accepted[0].Status != agentbridge.ResultCompleted {
		t.Fatal("internal cancellation rewrote retry outcome")
	}
}

func TestLateCancelPreservesPreparedUnknownTerminal(t *testing.T) {
	reporter := &cancellationDeliveryReporter{failTerminal: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "possibly committed answer"})
	cancelDeliveryClaim(a, task, tasks)
	// The mocked server has committed the first result and returns its same ACK.
	a.retryTerminalReports(context.Background(), tasks)
	if len(reporter.results) != 2 || !reflect.DeepEqual(reporter.results[0], reporter.results[1]) || reporter.accepted[0].Status != agentbridge.ResultCompleted || len(tasks) != 0 {
		t.Fatal("late cancellation changed an ambiguous already-prepared terminal")
	}
}

func TestRetainedLifecycleWithoutCancellationKeepsRetryOrder(t *testing.T) {
	reporter := &cancellationDeliveryReporter{failEvent: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	a.reportTaskEvent(context.Background(), task, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "answer"})
	if len(reporter.results) != 0 {
		t.Fatal("terminal overtook pending normal lifecycle")
	}
	a.retryEventReports(context.Background(), tasks)
	a.retryTerminalReports(context.Background(), tasks)
	if !reflect.DeepEqual(reporter.order, []string{"event:failed", "event:accepted", "terminal:accepted"}) || len(tasks) != 0 {
		t.Fatalf("normal retry order changed: %v", reporter.order)
	}
}

func TestOldClaimCancellationDoesNotSupersedeReplacementDelivery(t *testing.T) {
	old, cancelOld := context.WithCancel(t.Context())
	defer cancelOld()
	reporter := &cancellationDeliveryReporter{failEvent: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	a.reportTaskEvent(context.Background(), task, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	a.handleTaskCancel(context.Background(), &cancelMsg{taskID: task.taskID, claimContext: old, cause: context.Canceled}, tasks)
	if task.cancelCause != nil || len(task.pendingEvents) != 1 || len(task.supersededEvents) != 0 {
		t.Fatal("old cancellation touched replacement claim")
	}
	a.retryEventReports(context.Background(), tasks)
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "replacement answer"})
	if len(reporter.accepted) != 1 || reporter.accepted[0].Status != agentbridge.ResultCompleted || len(tasks) != 0 {
		t.Fatal("replacement did not complete")
	}
}

func TestCancellationBeforeFirstTerminalAttemptSupersedesBufferedCompletion(t *testing.T) {
	reporter := &cancellationDeliveryReporter{failEvent: true}
	a, task, tasks := cancellationDeliveryFixture(reporter)
	a.reportTaskEvent(context.Background(), task, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	deliverCancellationTestResult(a, task, tasks, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "buffered provider answer", Usage: agentbridge.Usage{PromptTokens: 9}})
	if len(reporter.results) != 0 || task.terminalResult.Status != agentbridge.ResultCompleted {
		t.Fatal("completion was not buffered behind lifecycle")
	}
	reporter.cancelling = true
	cancelDeliveryClaim(a, task, tasks)
	a.retryEventReports(context.Background(), tasks)
	a.retryTerminalReports(context.Background(), tasks)
	if len(reporter.accepted) != 1 || reporter.accepted[0].Status != agentbridge.ResultCancelled || len(tasks) != 0 {
		t.Fatalf("unattempted completion ignored cancellation: results=%+v remaining=%d", reporter.results, len(tasks))
	}
	if reporter.accepted[0].Usage.PromptTokens != 9 || reporter.accepted[0].Output != "buffered provider answer" {
		t.Fatal("buffered provenance/usage lost")
	}
}
