package supervisor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
)

type claimProofReporter struct {
	*reporterProbe
	scopes []*executionproof.Scope
	proofs []bool
	fail   bool
}

func (r *claimProofReporter) CompleteTask(ctx context.Context, _ string, _ agentbridge.Result) error {
	r.scopes = append(r.scopes, executionproof.ScopeFromContext(ctx))
	_, ok := executionproof.ProofFromContext(ctx)
	r.proofs = append(r.proofs, ok)
	if r.fail {
		return errors.New("fixture terminal failure")
	}
	return nil
}

func TestClaimProofWorkerFinishesBeforeNotification(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	work, err := scope.Tracker().BeginHostWork()
	if err != nil {
		t.Fatal(err)
	}
	a := &Actor{mailbox: make(chan envelope, 1), stoppedCh: make(chan struct{})}
	a.forwardActivation(taskActivationMsg{hostWork: work})
	<-a.mailbox
	proof, ok := scope.Tracker().Seal()
	if !ok || proof.StartedInvocations != 0 {
		t.Fatal("activation exposed pending host work")
	}
}

func TestClaimProofPendingPreparationAndVerification(t *testing.T) {
	for _, worker := range []string{"prepare", "verify", "correction"} {
		t.Run(worker, func(t *testing.T) {
			scope := executionproof.NewScope(true, nil)
			taskCtx, cancel := context.WithCancel(executionproof.ContextWithScope(context.Background(), scope))
			defer cancel()
			work, err := beginClaimHostWork(taskCtx)
			if err != nil {
				t.Fatal(err)
			}
			reporter := &claimProofReporter{reporterProbe: newReporterProbe(), fail: true}
			actor := &Actor{cfg: Config{Reporter: reporter}}
			task := &runningTask{taskID: "claim", ctx: taskCtx, cancel: cancel}
			inFlight := map[string]*runningTask{task.taskID: task}
			if actor.finishTaskWithResult(context.Background(), inFlight, task, agentbridge.Result{Status: agentbridge.ResultCancelled}) {
				t.Fatal("failed terminal unexpectedly ACKed")
			}
			work.Finished()
			reporter.fail = false
			if !actor.retryTerminalReports(context.Background(), inFlight) {
				t.Fatal("business completion blocked on unknown observation")
			}
			if len(reporter.proofs) != 2 || reporter.proofs[0] || reporter.proofs[1] || reporter.scopes[0] != scope || reporter.scopes[1] != scope {
				t.Fatal("pending/late worker changed original unknown proof")
			}
		})
	}
}

func TestClaimProofTerminalCarriesOriginalScope(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	claimCtx, cancel := context.WithCancel(executionproof.ContextWithScope(context.Background(), scope))
	defer cancel()
	attempt, err := scope.Tracker().BeginInvocation()
	if err != nil {
		t.Fatal(err)
	}
	attempt.Started()
	attempt.Exited(true)
	reporter := &claimProofReporter{reporterProbe: newReporterProbe()}
	actor := &Actor{cfg: Config{Reporter: reporter}}
	task := &runningTask{taskID: "claim", ctx: claimCtx, cancel: cancel}
	inFlight := map[string]*runningTask{task.taskID: task}
	unrelated := executionproof.ContextWithScope(context.Background(), executionproof.NewScope(false, nil))
	if !actor.finishTaskWithResult(unrelated, inFlight, task, agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}) {
		t.Fatal("terminal not ACKed")
	}
	if len(reporter.proofs) != 1 || !reporter.proofs[0] || reporter.scopes[0] != scope {
		t.Fatal("report context lost captured claim")
	}
	if _, err = scope.Tracker().BeginInvocation(); err == nil {
		t.Fatal("terminal seal allowed late continuation")
	}
}

type claimScopeEventReporter struct {
	*reporterProbe
	scopes []*executionproof.Scope
	fail   bool
}

func (r *claimScopeEventReporter) ReportEvent(ctx context.Context, _ string, _ agentbridge.Event) error {
	r.scopes = append(r.scopes, executionproof.ScopeFromContext(ctx))
	if r.fail {
		return errors.New("fixture event failure")
	}
	return nil
}

func TestClaimProofEventAndRetryCarryOriginalScope(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	task := &runningTask{taskID: "claim", ctx: executionproof.ContextWithScope(context.Background(), scope)}
	reporter := &claimScopeEventReporter{reporterProbe: newReporterProbe(), fail: true}
	actor := &Actor{cfg: Config{Reporter: reporter}}
	actor.reportTaskEvent(context.Background(), task, agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning})
	if len(task.pendingEvents) != 1 {
		t.Fatal("event fixture did not retain failure")
	}
	reporter.fail = false
	if !actor.flushNextEventReport(context.Background(), task) {
		t.Fatal("event retry failed")
	}
	if len(reporter.scopes) != 2 || reporter.scopes[0] != scope || reporter.scopes[1] != scope {
		t.Fatal("ordinary event/retry dropped original claim scope")
	}
}
