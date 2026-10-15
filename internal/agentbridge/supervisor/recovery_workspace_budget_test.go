package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestDefaultRecoveryBudgetDoesNotBecomeWorkspaceDeadline(t *testing.T) {
	policy, err := (*assignment.ProviderRecoveryPolicy)(nil).Normalized()
	if err != nil || policy.MaxElapsedSeconds != 30 {
		t.Fatalf("default policy=%+v %v", policy, err)
	}
	originalResolver, originalClone := resolveAssignmentGitExecutable, runAssignmentGitClone
	resolveAssignmentGitExecutable = func(string, string) (string, bool) { return "/usr/bin/git", true }
	observed := make(chan bool, 1)
	runAssignmentGitClone = func(ctx context.Context, _ string, _ []string) error {
		_, hasDeadline := ctx.Deadline()
		observed <- hasDeadline
		return nil
	}
	t.Cleanup(func() { resolveAssignmentGitExecutable = originalResolver; runAssignmentGitClone = originalClone })
	source := controlplane.NewMemorySource()
	source.Enqueue(slowPrepareTaskRequest())
	reporter := newReporterProbe()
	fake := process.NewFake()
	run := process.NewFakeRunning()
	fake.NextRunning = run
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")
	startRoutingSupervisor(t, Config{DaemonID: "test", Runtime: rt, Source: source, Reporter: reporter, Workdir: workdir.NewFSAdapter(t.TempDir())})
	select {
	case deadline := <-observed:
		if deadline {
			t.Fatal("default30-second provider budget became initial workspace deadline; a31-second clone cannot complete")
		}
	case <-time.After(time.Second):
		t.Fatal("clone missing")
	}
	select {
	case <-run.StartedRecv():
		completeTaskResultProcess(run)
	case <-time.After(time.Second):
		t.Fatal("provider not admitted")
	}
}

func TestDefaultRecoveryBudgetAllows31SecondAssignmentClone(t *testing.T) {
	originalResolver, originalClone := resolveAssignmentGitExecutable, runAssignmentGitClone
	resolveAssignmentGitExecutable = func(string, string) (string, bool) { return "/usr/bin/git", true }
	runAssignmentGitClone = func(ctx context.Context, _ string, _ []string) error {
		timer := time.NewTimer(31 * time.Second)
		defer timer.Stop()
		select {
		case <-timer.C:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	t.Cleanup(func() { resolveAssignmentGitExecutable = originalResolver; runAssignmentGitClone = originalClone })
	source := controlplane.NewMemorySource()
	source.Enqueue(slowPrepareTaskRequest())
	reporter := newReporterProbe()
	fake := process.NewFake()
	run := process.NewFakeRunning()
	fake.NextRunning = run
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")
	startRoutingSupervisor(t, Config{DaemonID: "test", Runtime: rt, Source: source, Reporter: reporter, Workdir: workdir.NewFSAdapter(t.TempDir())})
	select {
	case <-run.StartedRecv():
		completeTaskResultProcess(run)
	case res := <-reporter.results:
		t.Fatalf("normal31-second clone failed before dispatch: %+v", res)
	case <-time.After(35 * time.Second):
		t.Fatal("31-second clone never reached provider")
	}
}

type budgetConsumingReporter struct {
	*reporterProbe
	delayed bool
}

func (r *budgetConsumingReporter) ReportEvent(ctx context.Context, id string, event agentbridge.Event) error {
	if event.RecoveryContext != nil && event.RecoveryContext.State == "checking" && !r.delayed {
		r.delayed = true
		timer := time.NewTimer(900 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return r.reporterProbe.ReportEvent(ctx, id, event)
}

func TestWorkspacePreparationDoesNotResetConsumedRecoveryBudget(t *testing.T) {
	originalResolver, originalClone := resolveAssignmentGitExecutable, runAssignmentGitClone
	resolveAssignmentGitExecutable = func(string, string) (string, bool) { return "/usr/bin/git", true }
	runAssignmentGitClone = func(ctx context.Context, _ string, _ []string) error { return ctx.Err() }
	t.Cleanup(func() { resolveAssignmentGitExecutable = originalResolver; runAssignmentGitClone = originalClone })
	source := controlplane.NewMemorySource()
	req := slowPrepareTaskRequest()
	req.RecoveryPolicy = &assignment.ProviderRecoveryPolicy{MaxAttempts: 3, MaxElapsedSeconds: 1}
	source.Enqueue(req)
	reporter := &budgetConsumingReporter{reporterProbe: newReporterProbe()}
	p := &retryStartProcess{failures: 10, started: make(chan *process.FakeRunning, 1)}
	rt, _ := newRecoveryRuntime(t, string(req.Provider), "", p)
	startRoutingSupervisor(t, Config{DaemonID: "daemon", Runtime: rt, Source: source, Reporter: reporter, Workdir: workdir.NewFSAdapter(t.TempDir())})
	result := expectTaskResult(t, reporter.reporterProbe, "consumed recovery budget was reset")
	if p.count() != 1 || result.RecoveryContext == nil || result.RecoveryContext.Reason != "deadline_exceeded" {
		t.Fatalf("attempts=%d context=%+v", p.count(), result.RecoveryContext)
	}
}
