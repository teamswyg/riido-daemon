package supervisor

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSupervisorRetriesDetachedReportOnlyEventUntilAccepted(t *testing.T) {
	ctx := context.Background()
	reporter := newStateEventRetryReporter(1)
	actor := &Actor{cfg: Config{Reporter: reporter}}
	pending := []detachedReport{}
	msg := &taskEventMsg{taskID: "t-1", event: runningLifecycleEvent()}

	actor.handleTaskReportEvent(ctx, msg, map[string]*runningTask{}, &pending)
	expectStateEventAttempt(t, reporter, 1)
	expectDetachedReportCount(t, pending, 1)

	if !actor.retryDetachedReports(ctx, &pending) {
		t.Fatal("detached event report was not retried")
	}
	expectStateEventAttempt(t, reporter, 2)
	expectRetriedRunningEvent(t, reporter.reporterProbe)
	expectDetachedReportCount(t, pending, 0)
}

func TestSupervisorRetriesDetachedTerminalResultUntilAccepted(t *testing.T) {
	ctx := context.Background()
	reporter := newTerminalRetryReporter(1)
	actor := &Actor{cfg: Config{Reporter: reporter}}
	pending := []detachedReport{}
	msg := &taskResultMsg{taskID: "t-1", result: completedResult()}

	actor.handleTaskResult(ctx, msg, map[string]*runningTask{}, &pending)
	expectTerminalAttempt(t, reporter, 1)
	expectDetachedReportCount(t, pending, 1)

	if !actor.retryDetachedReports(ctx, &pending) {
		t.Fatal("detached terminal report was not retried")
	}
	expectTerminalAttempt(t, reporter, 2)
	expectTaskResultCompleted(t, reporter.reporterProbe)
	expectDetachedReportCount(t, pending, 0)
}

func TestSupervisorDoesNotRetainDetachedTextDelta(t *testing.T) {
	pending := []detachedReport{}
	ev := agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "stream"}

	retainDetachedReport(&pending, detachedEvent("t-1", ev))
	expectDetachedReportCount(t, pending, 0)
}

func runningLifecycleEvent() agentbridge.Event {
	return agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}
}

func completedResult() agentbridge.Result {
	return agentbridge.Result{Status: agentbridge.ResultCompleted}
}
