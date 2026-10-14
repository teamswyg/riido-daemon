package supervisor

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSupervisorRetriesWorkspaceWarningBeforeTerminalReport(t *testing.T) {
	reporter := newStateEventRetryReporter(1)
	actor := &Actor{cfg: Config{Reporter: reporter}}
	task := &runningTask{taskID: "t-1", events: failingWorkspaceEvents(t)}
	inFlight := map[string]*runningTask{task.taskID: task}
	ctx := context.Background()

	res := actor.recordTerminalResult(ctx, task, agentbridge.Result{Status: agentbridge.ResultCompleted})
	actor.finishTaskWithResult(ctx, inFlight, task, res)
	expectStateEventAttempt(t, reporter, 1)
	if len(inFlight) == 0 {
		t.Fatal("terminal report completed before workspace warning was retried")
	}

	actor.retryEventReports(ctx, inFlight)
	expectStateEventAttempt(t, reporter, 2)
	expectWorkspaceWarning(t, reporter.reporterProbe, "terminal result event append failed")
	actor.retryTerminalReports(ctx, inFlight)
	expectTaskResultCompleted(t, reporter.reporterProbe)
}
