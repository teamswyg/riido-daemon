package supervisor

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
)

func TestSupervisorRetriesPrepareWorkspaceWarning(t *testing.T) {
	reporter := newStateEventRetryReporter(1)
	actor := &Actor{
		cfg:       Config{Reporter: reporter},
		mailbox:   make(chan envelope, 1),
		stoppedCh: make(chan struct{}),
	}
	task := &runningTask{taskID: "t-1"}
	inFlight := map[string]*runningTask{task.taskID: task}
	ctx := context.Background()

	actor.appendWorkspaceEvent(ctx, task.taskID, failingWorkspaceEvents(t), ir.EventWorkdirCreated, "", nil)
	actor.handleMailboxMessage(ctx, expectReportOnlyEvent(t, actor), inFlight)
	expectStateEventAttempt(t, reporter, 1)

	actor.retryEventReports(ctx, inFlight)
	expectStateEventAttempt(t, reporter, 2)
	expectWorkspaceWarning(t, reporter.reporterProbe, "workspace event append failed")
}

func expectReportOnlyEvent(t *testing.T, actor *Actor) envelope {
	t.Helper()
	select {
	case msg := <-actor.mailbox:
		if msg.taskReport == nil {
			t.Fatal("mailbox event was not report-only")
		}
		return msg
	default:
		t.Fatal("report-only event was not forwarded")
		return envelope{}
	}
}
